import { writable } from 'svelte/store';

export const wsStatus = writable('DISCONNECTED'); // CONNECTED, CONNECTING, DISCONNECTED
export const incomingMessages = writable(null);

class WSClient {
  constructor() {
    this.ws = null;
    this.token = null;
    this.reconnectTimer = null;
  }

  connect(token) {
    if (!token) return;
    this.token = token;
    wsStatus.set('CONNECTING');

    const wsUrl = `ws://localhost:8080/ws?token=${token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      wsStatus.set('CONNECTED');
      console.log('🟢 WebSocket Connected');
      clearTimeout(this.reconnectTimer);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        incomingMessages.set(data);
      } catch (e) {
        console.error('Lỗi parse gói tin WS:', e);
      }
    };

    this.ws.onclose = () => {
      wsStatus.set('DISCONNECTED');
      console.log('🔴 WebSocket Disconnected, reconnecting in 3s...');
      this.reconnectTimer = setTimeout(() => this.connect(this.token), 3000);
    };

    this.ws.onerror = (err) => {
      console.error('WebSocket Error:', err);
      this.ws.close();
    };
  }

  send(event, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ event, payload }));
    } else {
      console.warn('WebSocket chưa sẵn sàng để gửi tin');
    }
  }

  disconnect() {
    clearTimeout(this.reconnectTimer);
    if (this.ws) this.ws.close();
  }
}

export const wsService = new WSClient();
