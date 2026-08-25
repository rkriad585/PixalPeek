declare module "/wails/runtime.js" {
  export type CancellablePromise<T> = Promise<T>;
  export const Call: {
    ByID(id: number, ...args: unknown[]): Promise<unknown>;
    ByName(name: string, ...args: unknown[]): Promise<unknown>;
  };
  export const Create: unknown;
  export const Events: {
    On(name: string, cb: (data: unknown) => void): () => void;
    Off(name: string): void;
    Emit(name: string, data?: unknown): void;
  };
  export class Window {
    constructor(name?: string);
    static Get(name?: string): Window;
    Close(): Promise<void>;
    Minimise(): Promise<void>;
    Maximise(): Promise<void>;
    UnMaximise(): Promise<void>;
    UnMinimise(): Promise<void>;
    ToggleMaximise(): Promise<void>;
    IsMaximised(): Promise<boolean>;
    IsMinimised(): Promise<boolean>;
    IsFullscreen(): Promise<boolean>;
    Fullscreen(): Promise<void>;
    UnFullscreen(): Promise<void>;
    ToggleFullscreen(): Promise<void>;
    Position(): Promise<{ x: number; y: number }>;
    Center(): Promise<void>;
    Show(): Promise<void>;
    Hide(): Promise<void>;
    Focus(): Promise<void>;
    Name(): Promise<string>;
    Size(): Promise<{ width: number; height: number }>;
    SetTitle(title: string): Promise<void>;
  }
  export const Browser: {
    OpenURL(url: string): Promise<void>;
  };
}
