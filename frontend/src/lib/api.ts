const BASE_URL = 'http://localhost:8080/api';

// fetchのラッパー (認証クッキーを常に送信)
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${BASE_URL}${path}`;
  const headers = new Headers(options.headers);
  
  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  // 常にCredentialsをincludeにすることで、HttpOnlyクッキー(session_token)を送信・保持する
  const config: RequestInit = {
    ...options,
    headers,
    credentials: 'include', 
  };

  const response = await fetch(url, config);

  if (!response.ok) {
    let errorMessage = 'Something went wrong';
    try {
      const errData = await response.json() as { error?: string };
      errorMessage = errData.error || errorMessage;
    } catch {
      // JSONではないエラーレスポンスの場合
      errorMessage = await response.text() || errorMessage;
    }
    throw new Error(errorMessage);
  }

  // 204 No Content の場合はパースしない
  if (response.status === 204) {
    return {} as T;
  }

  return response.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string, options?: RequestInit) => request<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body: unknown, options?: RequestInit) => 
    request<T>(path, { ...options, method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown, options?: RequestInit) => 
    request<T>(path, { ...options, method: 'PUT', body: JSON.stringify(body) }),
  delete: <T>(path: string, options?: RequestInit) => request<T>(path, { ...options, method: 'DELETE' }),
};
