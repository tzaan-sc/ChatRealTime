const API_BASE = 'http://localhost:8080/api';

let isRefreshing = false;
let refreshSubscribers = [];

function subscribeTokenRefresh(cb) {
  refreshSubscribers.push(cb);
}

function onRefreshed(newToken) {
  refreshSubscribers.forEach((cb) => cb(newToken));
  refreshSubscribers = [];
}

async function doRefreshToken() {
  const currentRefreshToken = localStorage.getItem('refresh_token');
  if (!currentRefreshToken) {
    throw new Error('Không có refresh token');
  }

  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: currentRefreshToken })
  });

  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Phiên đăng nhập đã hết hạn');
  }

  const newToken = data.data.token;
  const newRefreshToken = data.data.refresh_token;

  localStorage.setItem('token', newToken);
  if (newRefreshToken) {
    localStorage.setItem('refresh_token', newRefreshToken);
  }

  return newToken;
}

export async function apiRequest(endpoint, method = 'GET', body = null, token = null) {
  const activeToken = token || localStorage.getItem('token');
  const headers = { 'Content-Type': 'application/json' };
  if (activeToken) headers['Authorization'] = `Bearer ${activeToken}`;

  const config = { method, headers };
  if (body) config.body = JSON.stringify(body);

  const res = await fetch(`${API_BASE}${endpoint}`, config);
  const data = await res.json().catch(() => ({}));

  // Cơ chế Silent Token Refresh tự động khi Access Token hết hạn (401)
  if (res.status === 401 && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/refresh')) {
    const storedRefresh = localStorage.getItem('refresh_token');
    if (storedRefresh) {
      if (!isRefreshing) {
        isRefreshing = true;
        try {
          const newToken = await doRefreshToken();
          isRefreshing = false;
          onRefreshed(newToken);

          // Retry request ban đầu với Access Token mới
          headers['Authorization'] = `Bearer ${newToken}`;
          const retryRes = await fetch(`${API_BASE}${endpoint}`, config);
          const retryData = await retryRes.json().catch(() => ({}));
          if (!retryRes.ok) throw new Error(retryData.error || 'Có lỗi xảy ra');
          return retryData;
        } catch (err) {
          isRefreshing = false;
          refreshSubscribers = [];
          localStorage.removeItem('token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user');
          window.location.reload();
          throw err;
        }
      } else {
        // Đang có tiến trình refresh token chạy -> xếp hàng chờ
        return new Promise((resolve, reject) => {
          subscribeTokenRefresh(async (newToken) => {
            try {
              headers['Authorization'] = `Bearer ${newToken}`;
              const retryRes = await fetch(`${API_BASE}${endpoint}`, config);
              const retryData = await retryRes.json().catch(() => ({}));
              if (!retryRes.ok) throw new Error(retryData.error || 'Có lỗi xảy ra');
              resolve(retryData);
            } catch (err) {
              reject(err);
            }
          });
        });
      }
    }
  }

  if (!res.ok) {
    throw new Error(data.error || 'Có lỗi xảy ra');
  }
  return data;
}
