package services

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	winspool         = windows.NewLazySystemDLL("winspool.drv")
	procGetDefault   = winspool.NewProc("GetDefaultPrinterW")
	procSetDefault   = winspool.NewProc("SetDefaultPrinterW")
	procEnumPrinters = winspool.NewProc("EnumPrintersW")
	procOpenPrinter  = winspool.NewProc("OpenPrinterW")
	procGetPrinter   = winspool.NewProc("GetPrinterW")
	procClosePrinter = winspool.NewProc("ClosePrinter")
	procEnumJobs     = winspool.NewProc("EnumJobsW")
)

func coInit() error {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() == 1 {
			return nil
		}
		return fmt.Errorf("COM初始化失败: %v", err)
	}
	return nil
}

func getDefaultPrinter() string {
	var buf [256]uint16
	size := uint32(len(buf))
	procGetDefault.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	return windows.UTF16ToString(buf[:])
}

func setDefaultPrinter(name string) {
	ptr, _ := windows.UTF16PtrFromString(name)
	procSetDefault.Call(uintptr(unsafe.Pointer(ptr)))
}

// PrinterInfo 打印机信息
type PrinterInfo struct {
	Name     string `json:"name"`
	Status   int    `json:"status"`    // WMI PrinterStatus: 3=Idle, 4=Printing, 7=Offline
	JobCount int    `json:"job_count"` // 当前队列任务数
}

// CheckWPS 检测 WPS 是否安装
func CheckWPS() bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := coInit(); err != nil {
		return false
	}
	defer ole.CoUninitialize()
	unknown, err := oleutil.CreateObject("kwps.Application")
	if err != nil {
		return false
	}
	unknown.Release()
	return true
}

// CheckAcrobat 检测 Acrobat 是否安装（注册表检测，不启动进程）
func CheckAcrobat() bool {
	k, err := registry.OpenKey(registry.CLASSES_ROOT, `AcroExch.App`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

// QueryPrinters 通过 winspool.drv EnumPrintersW 查询本机打印机列表
func QueryPrinters() ([]PrinterInfo, error) {
	const PRINTER_ENUM_LOCAL = 2
	const PRINTER_ENUM_CONNECTIONS = 4

	var needed, returned uint32
	procEnumPrinters.Call(
		uintptr(PRINTER_ENUM_LOCAL|PRINTER_ENUM_CONNECTIONS),
		0, 2, 0, 0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)
	if needed == 0 {
		return nil, nil
	}

	buf := make([]byte, needed)
	ret, _, err := procEnumPrinters.Call(
		uintptr(PRINTER_ENUM_LOCAL|PRINTER_ENUM_CONNECTIONS),
		0, 2,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("EnumPrintersW failed: %v", err)
	}

	type PRINTER_INFO_2 struct {
		pServerName         *uint16
		pPrinterName        *uint16
		pShareName          *uint16
		pPortName           *uint16
		pDriverName         *uint16
		pComment            *uint16
		pLocation           *uint16
		pDevMode            uintptr
		pSepFile            *uint16
		pPrintProcessor     *uint16
		pDatatype           *uint16
		pParameters         *uint16
		pSecurityDescriptor uintptr
		Attributes          uint32
		Priority            uint32
		DefaultPriority     uint32
		StartTime           uint32
		UntilTime           uint32
		Status              uint32
		cJobs               uint32
		AveragePPM          uint32
	}

	const infoSize = unsafe.Sizeof(PRINTER_INFO_2{})
	printers := make([]PrinterInfo, 0, returned)
	for i := uint32(0); i < returned; i++ {
		info := (*PRINTER_INFO_2)(unsafe.Pointer(&buf[i*uint32(infoSize)]))
		// 将 PRINTER_STATUS_* 映射到 WMI 风格: 0=Idle→3, Printing→4, Offline→7
		status := 3
		const PRINTER_STATUS_PRINTING = 0x00000400
		const PRINTER_STATUS_OFFLINE = 0x00000080
		if info.Status&PRINTER_STATUS_OFFLINE != 0 {
			status = 7
		} else if info.Status&PRINTER_STATUS_PRINTING != 0 {
			status = 4
		}
		name := windows.UTF16PtrToString(info.pPrinterName)
		if strings.Contains(strings.ToLower(name), "pdf") {
			continue
		}
		if strings.Contains(strings.ToLower(name), "microsoft xps") {
			continue
		}
		printers = append(printers, PrinterInfo{
			Name:     name,
			Status:   status,
			JobCount: int(info.cJobs),
		})
	}
	return printers, nil
}

// JobInfo 打印任务信息
type JobInfo struct {
	JobID    uint32 `json:"job_id"`
	Document string `json:"document"`
	Status   uint32 `json:"status"`
}

// QueryPrinterJobs 查询指定打印机的队列任务
func QueryPrinterJobs(printerName string) ([]JobInfo, error) {
	ptr, err := windows.UTF16PtrFromString(printerName)
	if err != nil {
		return nil, err
	}
	var handle uintptr
	ret, _, e := procOpenPrinter.Call(uintptr(unsafe.Pointer(ptr)), uintptr(unsafe.Pointer(&handle)), 0)
	if ret == 0 {
		return nil, fmt.Errorf("OpenPrinter failed: %v", e)
	}
	defer procClosePrinter.Call(handle)

	var needed, returned uint32
	procEnumJobs.Call(handle, 0, 255, 1, 0, 0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)))
	if needed == 0 {
		return []JobInfo{}, nil
	}

	buf := make([]byte, needed)
	ret, _, e = procEnumJobs.Call(handle, 0, 255, 1,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)))
	if ret == 0 {
		return nil, fmt.Errorf("EnumJobs failed: %v", e)
	}

	type JOB_INFO_1 struct {
		JobId        uint32
		pPrinterName *uint16
		pMachineName *uint16
		pUserName    *uint16
		pDocument    *uint16
		pDatatype    *uint16
		pStatus      *uint16
		Status       uint32
		Priority     uint32
		Position     uint32
		TotalPages   uint32
		PagesPrinted uint32
		Submitted    [8]uint16 // SYSTEMTIME (16 bytes)
	}
	const jobSize = unsafe.Sizeof(JOB_INFO_1{})
	jobs := make([]JobInfo, 0, returned)
	for i := uint32(0); i < returned; i++ {
		j := (*JOB_INFO_1)(unsafe.Pointer(&buf[i*uint32(jobSize)]))
		jobs = append(jobs, JobInfo{
			JobID:    j.JobId,
			Document: windows.UTF16PtrToString(j.pDocument),
			Status:   j.Status,
		})
	}
	return jobs, nil
}

// PrintService 打印服务结构体
type PrintService struct{}

// OpenFile 通过软件打开文件
func (s *PrintService) OpenFile(filePath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取文件绝对路径失败: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(absPath))

	if err := coInit(); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	switch ext {
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		err = s.openOffice(absPath)
	case ".pdf":
		err = s.openPDF(absPath)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}
	return err
}

func (s *PrintService) openOffice(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	var appProgID string
	switch ext {
	case ".doc", ".docx":
		appProgID = "kwps.Application"
	case ".xls", ".xlsx":
		appProgID = "ket.Application"
	case ".ppt", ".pptx":
		appProgID = "kwpp.Application"
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	unknown, err := oleutil.CreateObject(appProgID)
	if err != nil {
		return fmt.Errorf("创建应用实例失败: %v", err)
	}
	defer unknown.Release()

	app, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("获取应用接口失败: %v", err)
	}
	defer app.Release()

	switch ext {
	case ".doc", ".docx":
		docs := oleutil.MustGetProperty(app, "Documents").ToIDispatch()
		defer docs.Release()
		doc := oleutil.MustCallMethod(docs, "Open", filePath).ToIDispatch()
		defer doc.Release()
	case ".xls", ".xlsx":
		workbooks := oleutil.MustGetProperty(app, "WorkBooks").ToIDispatch()
		defer workbooks.Release()
		wb := oleutil.MustCallMethod(workbooks, "Open", filePath).ToIDispatch()
		defer wb.Release()
	case ".ppt", ".pptx":
		presentations := oleutil.MustGetProperty(app, "Presentations").ToIDispatch()
		defer presentations.Release()
		ppt := oleutil.MustCallMethod(presentations, "Open", filePath).ToIDispatch()
		defer ppt.Release()
	}
	oleutil.MustPutProperty(app, "Visible", true)
	return nil
}

func (s *PrintService) openPDF(filePath string) error {
	unknown, err := oleutil.CreateObject("AcroExch.App")
	if err != nil {
		return fmt.Errorf("创建PDF应用实例失败: %v", err)
	}
	defer unknown.Release()

	pdf, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("获取PDF接口失败: %v", err)
	}
	defer pdf.Release()

	unknownDoc, err := oleutil.CreateObject("AcroExch.AVDoc")
	if err != nil {
		return fmt.Errorf("创建PDF文档实例失败: %v", err)
	}
	defer unknownDoc.Release()

	pdfDoc, err := unknownDoc.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("获取PDF文档接口失败: %v", err)
	}
	defer pdfDoc.Release()

	if _, err = oleutil.CallMethod(pdfDoc, "Open", filePath, ""); err != nil {
		return fmt.Errorf("打开PDF文档失败: %v", err)
	}
	if _, err = oleutil.CallMethod(pdf, "Show"); err != nil {
		return fmt.Errorf("打开PDF应用失败: %v", err)
	}
	return nil
}

// PrintFile 处理文件打印，printerName 为空时使用系统默认打印机
func (s *PrintService) PrintFile(filePath, printerName string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取文件绝对路径失败: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(absPath))

	if err := coInit(); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	switch ext {
	case ".doc", ".docx":
		err = s.printWord(absPath, printerName)
	case ".pdf":
		err = s.printPDF(absPath, printerName)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}
	return err
}

func (s *PrintService) printWord(filePath, printerName string) error {
	unknown, err := oleutil.CreateObject("kwps.Application")
	if err != nil {
		return fmt.Errorf("创建Word应用实例失败: %v", err)
	}
	defer unknown.Release()

	word, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("获取Word接口失败: %v", err)
	}
	defer word.Release()

	if printerName != "" {
		oleutil.MustPutProperty(word, "ActivePrinter", printerName)
	}

	docs := oleutil.MustGetProperty(word, "Documents").ToIDispatch()
	defer docs.Release()
	doc := oleutil.MustCallMethod(docs, "Open", filePath).ToIDispatch()
	defer doc.Release()

	if _, err = oleutil.CallMethod(doc, "PrintOut"); err != nil {
		return fmt.Errorf("打印文档失败: %v", err)
	}
	if _, err = oleutil.CallMethod(doc, "Close"); err != nil {
		return fmt.Errorf("关闭文档失败: %v", err)
	}
	if _, err = oleutil.CallMethod(word, "Quit"); err != nil {
		return fmt.Errorf("退出Word应用失败: %v", err)
	}
	return nil
}

func (s *PrintService) printPDF(filePath, printerName string) error {
	if printerName != "" {
		prev := getDefaultPrinter()
		setDefaultPrinter(printerName)
		defer setDefaultPrinter(prev)
	}

	unknown, err := oleutil.CreateObject("AcroExch.AVDoc")
	if err != nil {
		return fmt.Errorf("创建PDF应用实例失败: %v", err)
	}
	defer unknown.Release()

	pdf, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("获取PDF接口失败: %v", err)
	}
	defer pdf.Release()

	if _, err = oleutil.CallMethod(pdf, "Open", filePath, ""); err != nil {
		return fmt.Errorf("打开PDF文档失败: %v", err)
	}

	pdDoc := oleutil.MustCallMethod(pdf, "GetPDDoc").ToIDispatch()
	defer pdDoc.Release()

	pages := oleutil.MustCallMethod(pdDoc, "GetNumPages").Val
	if pages <= 0 {
		return fmt.Errorf("PDF页数无效: %d", pages)
	}

	if _, err = oleutil.CallMethod(pdf, "PrintPages", 0, pages-1, 2, 1, 1); err != nil {
		return fmt.Errorf("打印PDF文档失败: %v", err)
	}
	if _, err = oleutil.CallMethod(pdf, "Close", true); err != nil {
		return fmt.Errorf("关闭PDF文档失败: %v", err)
	}
	return nil
}
