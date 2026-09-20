import { writable } from 'svelte/store';
import { apiRequest } from '../services/api';

const storedToken = localStorage.getItem('token');
const storedUser = localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')) : null;

export const token = writable(storedToken);
export const currentUser = writable(storedUser);

export function loginSuccess(userVal, tokenVal) {
  localStorage.setItem('token', tokenVal);
  localStorage.setItem('user', JSON.stringify(userVal));
  token.set(tokenVal);
  currentUser.set(userVal);
}

export function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  token.set(null);
  currentUser.set(null);
}
