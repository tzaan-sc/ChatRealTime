# Hướng Dẫn Chi Tiết Giai Đoạn 9: Cuộc Gọi Thoại & Video Call 1-1 (WebRTC Realtime)
> **Dành cho người mới bắt đầu** | Hệ điều hành: Windows | Stack: WebRTC Peer-to-Peer + Golang WebSocket Signaling + Svelte 5 + Google STUN

Giai đoạn này sẽ mang đến tính năng "xịn xò" nhất của các ứng dụng chat hiện đại: **Cuộc gọi Audio & Video Call 1-1 trực tiếp thời gian thực** chất lượng cao và độ trễ cực thấp:
1. 📞 **Gọi thoại (Voice Call) & 📹 Gọi hình ảnh (Video Call):** Bấm nút gọi trên thanh chat $\to$ Trình duyệt đối phương đổ chuông kêu vang.
2. 🔄 **Cơ chế Báo hiệu (Signaling qua WebSocket Hub hiện có):** 
   - Không cần dựng thêm server thứ 3! Tận dụng chính kết nối WebSocket Go + Redis hiện có để trao đổi thông tin kết nối (SDP Offer, SDP Answer, ICE Candidates).
3. 🌐 **Kết nối Trực tiếp Ngang hàng (P2P Peer-to-Peer):**
   - Sau khi bắt tay thành công qua STUN server miễn phí của Google (`stun:stun.l.google.com:19302`), luồng âm thanh và hình ảnh truyền trực tiếp giữa 2 máy tính mà **không tốn băng thông máy chủ**!
4. 🖥️ **Giao diện Gọi điện Sang trọng (Call Modal):**
   - Màn hình chuông reo đón nhận cuộc gọi (Chấp nhận / Từ chối).
   - Màn hình đàm thoại Video lớn với ô camera nhỏ của bản thân góc dưới (Picture-in-Picture).
   - Các nút điều khiển: Bật/Tắt Mic 🎙️, Bật/Tắt Camera 📷, Chia sẻ màn hình 🖥️, Kết thúc cuộc gọi 🛑.

---

## 📋 Checklist Tiến Độ Giai Đoạn 9

- [x] **1. Bổ sung Signaling Events vào WebSocket Hub (Golang):**
  - [x] `call:request`: Người gọi bắt đầu cuộc gọi (chứa `receiver_id`, `call_type`: "video"|"audio").
  - [x] `call:accept`: Người nhận đồng ý nghe máy.
  - [x] `call:reject`: Người nhận từ chối cuộc gọi.
  - [x] `call:offer`: Trao đổi thông số SDP Offer.
  - [x] `call:answer`: Trao đổi thông số SDP Answer.
  - [x] `call:ice_candidate`: Trao đổi địa chỉ mạng ICE Candidate qua STUN.
  - [x] `call:hangup`: Kết thúc cuộc gọi từ một trong hai phía.
- [x] **2. Xây dựng Trình Quản lý WebRTC trên Frontend (`webrtc.js`):**
  - [x] Khởi tạo `RTCPeerConnection` với cấu hình ICE Google STUN Server.
  - [x] Lấy luồng Media (`navigator.mediaDevices.getUserMedia`).
  - [x] Gán Local Stream và Remote Stream vào các thẻ `<video>` HTML5.
- [x] **3. Xây dựng Component Giao diện Cuộc gọi (`CallModal.svelte`):**
  - [x] Giao diện cuộc gọi đến kèm âm thanh chuông reo (Web Audio API).
  - [x] Khung hình Video toàn màn hình chuẩn điện ảnh Cyber Glassmorphism.
  - [x] Bảng nút chức năng điều khiển âm thanh, hình ảnh và cúp máy.
- [x] **4. Tích hợp nút Gọi vào `ChatArea.svelte`:**
  - [x] Icon Gọi thoại 📞 và Gọi Video 📹 trên thanh Header của cuộc trò chuyện.
- [x] **5. Kiểm thử toàn diện các kịch bản gọi điện P2P**.

---

## Mục lục
1. [Bước 1: Cơ chế Báo hiệu (Signaling Hub trong Golang)](#buoc-1-signaling-go)
2. [Bước 2: Xây dựng Module Quản lý WebRTC (`webrtc.js`)](#buoc-2-webrtc-module)
3. [Bước 3: Tạo Giao diện Cuộc Gọi Cao Cấp (`CallModal.svelte`)](#buoc-3-call-modal)
4. [Bước 4: Tích hợp vào Khung Chat & App Svelte](#buoc-4-integration)
5. [Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết](#buoc-5-kiem-thu-thuc-te)

---

<a name="buoc-1-signaling-go"></a>
## Bước 1: Cơ chế Báo hiệu (Signaling Hub trong Golang)

Trong WebRTC, hai trình duyệt không thể tự biết IP của nhau nếu không có bên trung gian. Backend Golang đóng vai trò làm "Người đưa thư" (Signaling Server).

Trong `backend/internal/websocket/hub.go`, chỉ cần thêm case chuyển tiếp các gói tin gọi điện từ người gửi tới người nhận:

```go
case "call:request", "call:accept", "call:reject", "call:offer", "call:answer", "call:ice_candidate", "call:hangup":
    // Đóng gói và Publish vào Redis channel của người nhận
    receiverChannel := fmt.Sprintf("user:%s", targetUserID)
    h.redisClient.Publish(h.ctx, receiverChannel, messageBytes)
```

> 💡 **Ưu điểm vượt trội:** Backend chỉ đóng vai trò chuyển tiếp vài KB thông tin bắt tay (Signaling). Toàn bộ video nặng hàng trăm Megabyte sẽ chạy thẳng P2P giữa hai máy tính của người dùng!

---

<a name="buoc-2-webrtc-module"></a>
## Bước 2: Xây dựng Module Quản lý WebRTC (`webrtc.js`)

Tạo file `frontend/src/lib/utils/webrtc.js`:

```javascript
// Cấu hình STUN server miễn phí của Google để vượt qua NAT mạng gia đình
const rtcConfig = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' }
  ]
};

export class WebRTCManager {
  constructor(onRemoteStream, onIceCandidate) {
    this.peerConnection = null;
    this.localStream = null;
    this.onRemoteStream = onRemoteStream;
    this.onIceCandidate = onIceCandidate;
  }

  async initLocalStream(video = true, audio = true) {
    this.localStream = await navigator.mediaDevices.getUserMedia({ video, audio });
    return this.localStream;
  }

  createPeerConnection() {
    this.peerConnection = new RTCPeerConnection(rtcConfig);

    // Gắn local tracks vào peer connection
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => {
        this.peerConnection.addTrack(track, this.localStream);
      });
    }

    // Khi nhận được luồng video từ đối phương
    this.peerConnection.ontrack = (event) => {
      if (this.onRemoteStream && event.streams[0]) {
        this.onRemoteStream(event.streams[0]);
      }
    };

    // Khi tìm thấy ICE candidate
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate && this.onIceCandidate) {
        this.onIceCandidate(event.candidate);
      }
    };

    return this.peerConnection;
  }

  async createOffer() {
    const offer = await this.peerConnection.createOffer();
    await this.peerConnection.setLocalDescription(offer);
    return offer;
  }

  async handleOfferAndCreateAnswer(offer) {
    await this.peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await this.peerConnection.createAnswer();
    await this.peerConnection.setLocalDescription(answer);
    return answer;
  }

  async handleAnswer(answer) {
    await this.peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
  }

  async addIceCandidate(candidate) {
    if (this.peerConnection) {
      await this.peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
    }
  }

  close() {
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => track.stop());
      this.localStream = null;
    }
    if (this.peerConnection) {
      this.peerConnection.close();
      this.peerConnection = null;
    }
  }
}
```

---

<a name="buoc-3-call-modal"></a>
## Bước 3: Tạo Giao diện Cuộc Gọi Cao Cấp (`CallModal.svelte`)

Giao diện gồm 2 trạng thái:
1. **Trạng thái chuông reo (Incoming/Outgoing Call):**
   - Avatar người gọi đập theo nhịp chuông (pulse animation).
   - Nút Xanh (Nghe máy) & Nút Đỏ (Từ chối).
2. **Trạng thái đang gọi (Active Call):**
   - Khung hình đối phương phủ tràn màn hình sắc nét.
   - Khung hình của chính mình nhỏ gọn ở góc dưới bên phải.
   - Đồng hồ đếm thời gian cuộc gọi `02:45`.
   - Nút Mute Mic, Tắt Camera, Cúp máy.

---

<a name="buoc-4-integration"></a>
## Bước 4: Tích hợp vào Khung Chat & App Svelte

Trên thanh Header của `ChatArea.svelte`, thêm 2 nút:
```svelte
<div class="call-actions">
  <button class="call-btn" on:click={() => startCall('audio')} title="Gọi thoại">
    <Phone size={18} />
  </button>
  <button class="call-btn primary" on:click={() => startCall('video')} title="Gọi Video">
    <Video size={18} />
  </button>
</div>
```

---

<a name="buoc-5-kiem-thu-thuc-te"></a>
## Bước 5: Kịch bản Kiểm thử Thực tế Chi tiết

| Kịch bản | Thao tác thực hiện | Kết quả mong đợi |
| :--- | :--- | :--- |
| 📹 **Bắt đầu gọi Video** | Alex bấm icon Video Call 📹 ở khung chat với Bob | Màn hình của Bob lập tức bật chuông reo báo "Alex đang gọi video cho bạn...", màn hình Alex hiện "Đang kết nối...". |
| ✅ **Chấp nhận cuộc gọi** | Bob bấm nút xanh "Trả lời" | Khung hình video của cả 2 bên kết nối thành công, nhìn thấy và nghe thấy nhau mượt mà với độ trễ < 50ms! |
| 🛑 **Từ chối / Cúp máy** | Một trong hai người bấm nút cúp máy màu đỏ | Cuộc gọi kết thúc ngay lập tức, camera và micro được giải phóng, màn hình trở về khung chat. |
| 🎙️ **Tắt / Bật Micro** | Bấm nút Mute Mic trong cuộc gọi | Âm thanh tạm ngắt, đối phương không nghe thấy tiếng; bấm lại để mở micro bình thường. |
| 📷 **Tắt / Bật Camera** | Bấm nút Camera Toggle | Hình ảnh chuyển sang hiển thị Avatar đại diện; bật lại để tiếp tục truyền video. |

---

### 🎉 Hoàn thành Giai đoạn 9!
Ứng dụng chat của bạn đã sánh ngang với các ứng dụng nhắn tin hàng đầu thế giới với tính năng **Đàm thoại Video Realtime WebRTC P2P**!
