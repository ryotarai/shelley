import type * as Monaco from "monaco-editor";
import { withBasePath } from "./paths";

// Global Monaco instance - loaded lazily, shared across components
let monacoInstance: typeof Monaco | null = null;
let monacoLoadPromise: Promise<typeof Monaco> | null = null;

export function loadMonaco(): Promise<typeof Monaco> {
  if (monacoInstance) {
    return Promise.resolve(monacoInstance);
  }
  if (monacoLoadPromise) {
    return monacoLoadPromise;
  }

  monacoLoadPromise = (async () => {
    // Configure Monaco environment for web workers before importing
    const monacoEnv: Monaco.Environment = {
      getWorkerUrl: () => withBasePath("/editor.worker.js"),
    };
    (self as Window).MonacoEnvironment = monacoEnv;

    // Load Monaco CSS if not already loaded
    const monacoCssPath = withBasePath("/monaco-editor.css");
    if (!document.querySelector(`link[href="${monacoCssPath}"]`)) {
      const link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = monacoCssPath;
      document.head.appendChild(link);
    }

    // Load Monaco from our local bundle (runtime URL, cast to proper types)
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore - dynamic runtime URL import
    const monaco = (await import(/* @vite-ignore */ withBasePath("/monaco-editor.js"))) as typeof Monaco;
    monacoInstance = monaco;
    return monacoInstance;
  })();

  return monacoLoadPromise;
}
