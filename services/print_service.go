package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// PrintService 打印服务结构体
type PrintService struct{}

// PrintFile 通过软件打开文件
func (s *PrintService) OpenFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取文件绝对路径失败: %v", err)
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(absPath))

	// 初始化COM
	err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		return fmt.Errorf("COM初始化失败: %v", err)
	}
	defer ole.CoUninitialize()

	// 根据文件类型选择打开方式
	switch ext {
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		err = s.openOffice(absPath)
	case ".pdf":
		err = s.openPDF(absPath)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	if err != nil {
		return err
	}

	return nil
}

// openWord 打开Word文档
func (s *PrintService) openOffice(filePath string) error {
	// 根据文件扩展名选择应用程序
	ext := strings.ToLower(filepath.Ext(filePath))
	var appProgID string
	switch ext {
	case ".doc", ".docx":
		appProgID = "kwps.Application" // Word
	case ".xls", ".xlsx":
		appProgID = "ket.Application" // Excel
	case ".ppt", ".pptx":
		appProgID = "kwpp.Application" // PPT
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 创建应用实例
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

	// 打开文件
	switch ext {
	case ".doc", ".docx":
		docs := oleutil.MustGetProperty(app, "Documents").ToIDispatch()
		doc := oleutil.MustCallMethod(docs, "Open", filePath).ToIDispatch()
		defer doc.Release()
	case ".xls", ".xlsx":
		workbooks := oleutil.MustGetProperty(app, "WorkBooks").ToIDispatch()
		wb := oleutil.MustCallMethod(workbooks, "Open", filePath).ToIDispatch()
		defer wb.Release()
	case ".ppt", ".pptx":
		presentations := oleutil.MustGetProperty(app, "Presentations").ToIDispatch()
		ppt := oleutil.MustCallMethod(presentations, "Open", filePath).ToIDispatch()
		defer ppt.Release()
	}

	// 设置应用可见
	oleutil.MustPutProperty(app, "Visible", true)

	return nil
}

// openPDF 打开PDF文档
func (s *PrintService) openPDF(filePath string) error {
	// 创建PDF应用实例
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

	// 创建PDF文档实例
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

	// 打开PDF文档
	_, err = oleutil.CallMethod(pdfDoc, "Open", filePath, "")
	if err != nil {
		return fmt.Errorf("打开PDF文档失败: %v", err)
	}

	// 打开PDF应用
	_, err = oleutil.CallMethod(pdf, "Show")
	if err != nil {
		return fmt.Errorf("打开PDF应用失败: %v", err)
	}

	return nil
}

// PrintFile 处理文件打印
func (s *PrintService) PrintFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取文件绝对路径失败: %v", err)
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(absPath))

	// 初始化COM
	err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		return fmt.Errorf("COM初始化失败: %v", err)
	}
	defer ole.CoUninitialize()

	// 根据文件类型选择打印方式
	switch ext {
	case ".doc", ".docx":
		err = s.printWord(absPath)
	case ".pdf":
		err = s.printPDF(absPath)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	if err != nil {
		return err
	}

	// 打印完成后删除文件
	// err = os.Remove(absPath)
	// if err != nil {
	// 	return fmt.Errorf("删除文件失败: %v", err)
	// }

	return nil
}

// printWord 打印Word文档
func (s *PrintService) printWord(filePath string) error {
	// 创建Word应用实例
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

	// 打开文档
	docs := oleutil.MustGetProperty(word, "Documents").ToIDispatch()
	doc := oleutil.MustCallMethod(docs, "Open", filePath).ToIDispatch()
	defer doc.Release()

	// 打印文档
	_, err = oleutil.CallMethod(doc, "PrintOut")
	if err != nil {
		return fmt.Errorf("打印文档失败: %v", err)
	}

	// 关闭文档
	_, err = oleutil.CallMethod(doc, "Close")
	if err != nil {
		return fmt.Errorf("关闭文档失败: %v", err)
	}

	// 退出Word应用
	_, err = oleutil.CallMethod(word, "Quit")
	if err != nil {
		return fmt.Errorf("退出Word应用失败: %v", err)
	}

	return nil
}

// printPDF 打印PDF文档
func (s *PrintService) printPDF(filePath string) error {
	// 创建PDF应用实例
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

	// 打开PDF文档
	_, err = oleutil.CallMethod(pdf, "Open", filePath, "")
	if err != nil {
		return fmt.Errorf("打开PDF文档失败: %v", err)
	}

	// 获取PDDoc对象
	pdDoc := oleutil.MustCallMethod(pdf, "GetPDDoc").ToIDispatch()
	defer pdDoc.Release()

	// 获取页数
	pages := oleutil.MustCallMethod(pdDoc, "GetNumPages").Val

	// 打印文档
	_, err = oleutil.CallMethod(pdf, "PrintPages", 0, pages-1, 2, 1, 1)
	if err != nil {
		return fmt.Errorf("打印PDF文档失败: %v", err)
	}

	// 关闭文档
	_, err = oleutil.CallMethod(pdf, "Close", true)
	if err != nil {
		return fmt.Errorf("关闭PDF文档失败: %v", err)
	}

	return nil
}
