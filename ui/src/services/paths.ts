function normalizedBasePath(): string {
  const basePath = window.__SHELLEY_INIT__?.base_path || "/";
  if (basePath === "/" || basePath === "") {
    return "";
  }
  return basePath.endsWith("/") ? basePath.slice(0, -1) : basePath;
}

export function withBasePath(path: string): string {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    const url = new URL(path);
    url.pathname = withBasePath(url.pathname);
    return url.toString();
  }
  if (!path.startsWith("/")) {
    throw new Error(`Path must start with '/': ${path}`);
  }
  const basePath = normalizedBasePath();
  if (basePath === "") {
    return path;
  }
  return path === "/" ? `${basePath}/` : `${basePath}${path}`;
}

export function apiPath(path: string): string {
  if (!path.startsWith("/")) {
    throw new Error(`API path must start with '/': ${path}`);
  }
  return withBasePath(`/api${path}`);
}

export function stripBasePath(pathname: string): string {
  const basePath = normalizedBasePath();
  if (basePath === "") {
    return pathname;
  }
  if (pathname === basePath) {
    return "/";
  }
  if (pathname.startsWith(basePath + "/")) {
    return pathname.slice(basePath.length);
  }
  return pathname;
}
