export async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('pt_token');

  const headers = new Headers(options.headers);
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const config: RequestInit = {
    ...options,
    headers,
  };

  const response = await fetch(url, config);

  if (response.status === 401) {
    // Handle unauthorized - maybe redirect to login
    localStorage.removeItem('pt_token');
    window.location.href = '/login';
  }

  return response;
}
