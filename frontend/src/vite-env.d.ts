/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// 声明@novnc/novnc模块
// declare module '@novnc/novnc/lib/rfb' {
  
//   export class RFB {
//     constructor(target: HTMLElement | string, url: string, options?: any);
//     disconnect(): void;
//     connect(url: string, options?: any): void;
//     getDisplay(): any;
//     setScale(scale: number): void;
//     addEventListener(event: string, callback: (e: any) => void): void;
//     removeEventListener(event: string, callback: (e: any) => void): void;
//   }
// }


declare module '@novnc/novnc/lib/rfb' {
  import { EventEmitter } from 'events';
  
  export default class RFB extends EventEmitter {
    constructor(
      container: HTMLElement, 
      url: string, 
      options?: {
        credentials?: { password: string },
        shared?: boolean,
        clipViewport?: boolean,
        scaleViewport?: boolean,
        resizeSession?: boolean
        encrypt?: boolean,
      }
    );
    
    disconnect(): void;
    setViewportScale(scale: number): void;
    addEventListener(event: string, callback: (e: any) => void): void;
  }
}