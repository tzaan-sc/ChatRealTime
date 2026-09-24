import { writable } from 'svelte/store';

const storedToken = localStorage.getItem('token');
const storedRefreshToken = localStorage.getItem('refresh_token');
const storedUser = localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')) : null;

export const token = writable(storedToken);
export const refreshToken = writable(storedRefreshToken);
export const currentUser = writable(storedUser);
export const securityAlert = writable(null); // Lưu thông báo nếu phát hiện thiết bị lạ khi đăng nhập

export function loginSuccess(userVal, tokenVal, refreshTokenVal = null, alertMsg = null) {
  localStorage.setItem('token', tokenVal);
  token.set(tokenVal);

  if (refreshTokenVal) {
    localStorage.setItem('refresh_token', refreshTokenVal);
    refreshToken.set(refreshTokenVal);
  }

  localStorage.setItem('user', JSON.stringify(userVal));
  currentUser.set(userVal);

  if (alertMsg) {
    securityAlert.set(alertMsg);
  }
}

export function updateTokens(newToken, newRefreshToken) {
  if (newToken) {
    localStorage.setItem('token', newToken);
    token.set(newToken);
  }
  if (newRefreshToken) {
    localStorage.setItem('refresh_token', newRefreshToken);
    refreshToken.set(newRefreshToken);
  }
}

export function logout() {
  const currentRefresh = localStorage.getItem('refresh_token');
  if (currentRefresh) {
    fetch('http://localhost:8080/api/auth/logout', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: currentRefresh })
    }).catch(() => {});
  }

  localStorage.removeItem('token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('user');
  token.set(null);
  refreshToken.set(null);
  currentUser.set(null);
  securityAlert.set(null);
}

