// Cấu hình STUN servers của Google để vượt qua NAT
const rtcConfig = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
    { urls: 'stun:stun2.l.google.com:19302' }
  ]
};

export class WebRTCManager {
  constructor(options = {}) {
    this.peerConnection = null;
    this.localStream = null;
    this.remoteStream = null;
    this.screenStream = null;
    this.originalVideoTrack = null;

    this.onRemoteStream = options.onRemoteStream || (() => {});
    this.onIceCandidate = options.onIceCandidate || (() => {});
    this.onConnectionStateChange = options.onConnectionStateChange || (() => {});
  }

  // Khởi tạo luồng media từ Micro & Camera của máy
  async initLocalStream(callType = 'video') {
    try {
      const constraints = {
        audio: true,
        video: callType === 'video' ? { width: { ideal: 1280 }, height: { ideal: 720 } } : false
      };
      this.localStream = await navigator.mediaDevices.getUserMedia(constraints);
      return this.localStream;
    } catch (err) {
      console.error('Lỗi truy cập camera/micro:', err);
      // Fallback: nếu gọi video mà không có webcam thì thử chỉ lấy mic
      if (callType === 'video') {
        try {
          this.localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
          return this.localStream;
        } catch (audioErr) {
          throw audioErr;
        }
      }
      throw err;
    }
  }

  // Khởi tạo RTCPeerConnection
  createPeerConnection() {
    this.peerConnection = new RTCPeerConnection(rtcConfig);

    // Gắn local tracks vào peer connection
    if (this.localStream) {
      this.localStream.getTracks().forEach((track) => {
        this.peerConnection.addTrack(track, this.localStream);
      });
    }

    // Khi nhận được luồng stream từ đối phương
    this.peerConnection.ontrack = (event) => {
      if (event.streams && event.streams[0]) {
        this.remoteStream = event.streams[0];
        this.onRemoteStream(this.remoteStream);
      }
    };

    // Khi tìm thấy ICE candidate
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.onIceCandidate(event.candidate);
      }
    };

    // Theo dõi trạng thái kết nối
    this.peerConnection.onconnectionstatechange = () => {
      if (this.peerConnection) {
        this.onConnectionStateChange(this.peerConnection.connectionState);
      }
    };

    return this.peerConnection;
  }

  // Tạo SDP Offer (dành cho người gọi)
  async createOffer() {
    if (!this.peerConnection) this.createPeerConnection();
    const offer = await this.peerConnection.createOffer({
      offerToReceiveAudio: true,
      offerToReceiveVideo: true
    });
    await this.peerConnection.setLocalDescription(offer);
    return offer;
  }

  // Nhận SDP Offer và tạo SDP Answer (dành cho người nghe)
  async handleOfferAndCreateAnswer(offer) {
    if (!this.peerConnection) this.createPeerConnection();
    await this.peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await this.peerConnection.createAnswer();
    await this.peerConnection.setLocalDescription(answer);
    return answer;
  }

  // Nhận SDP Answer (dành cho người gọi)
  async handleAnswer(answer) {
    if (this.peerConnection) {
      await this.peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
    }
  }

  // Thêm ICE candidate nhận được từ signaling
  async addIceCandidate(candidate) {
    if (this.peerConnection && candidate) {
      try {
        await this.peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
      } catch (err) {
        console.warn('Lỗi thêm ICE Candidate:', err);
      }
    }
  }

  // Bật/Tắt Micro
  toggleAudio(enabled) {
    if (this.localStream) {
      this.localStream.getAudioTracks().forEach((t) => {
        t.enabled = enabled;
      });
    }
  }

  // Bật/Tắt Camera
  toggleVideo(enabled) {
    if (this.localStream) {
      this.localStream.getVideoTracks().forEach((t) => {
        t.enabled = enabled;
      });
    }
  }

  // Bắt đầu chia sẻ màn hình
  async startScreenShare() {
    try {
      this.screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
      const screenTrack = this.screenStream.getVideoTracks()[0];

      if (this.peerConnection) {
        const senders = this.peerConnection.getSenders();
        const videoSender = senders.find((s) => s.track && s.track.kind === 'video');

        if (videoSender) {
          this.originalVideoTrack = videoSender.track;
          await videoSender.replaceTrack(screenTrack);
        } else {
          this.peerConnection.addTrack(screenTrack, this.screenStream);
        }
      }

      screenTrack.onended = () => {
        this.stopScreenShare();
      };

      return this.screenStream;
    } catch (err) {
      console.warn('Hủy hoặc lỗi chia sẻ màn hình:', err);
      return null;
    }
  }

  // Dừng chia sẻ màn hình và khôi phục camera
  async stopScreenShare() {
    if (this.screenStream) {
      this.screenStream.getTracks().forEach((t) => t.stop());
      this.screenStream = null;
    }

    if (this.peerConnection && this.originalVideoTrack) {
      const senders = this.peerConnection.getSenders();
      const videoSender = senders.find((s) => s.track && s.track.kind === 'video');
      if (videoSender) {
        await videoSender.replaceTrack(this.originalVideoTrack);
      }
      this.originalVideoTrack = null;
    }
  }

  // Dọn dẹp kết nối và dừng tất cả camera/micro
  close() {
    if (this.screenStream) {
      this.screenStream.getTracks().forEach((t) => t.stop());
      this.screenStream = null;
    }
    if (this.localStream) {
      this.localStream.getTracks().forEach((t) => t.stop());
      this.localStream = null;
    }
    if (this.peerConnection) {
      this.peerConnection.close();
      this.peerConnection = null;
    }
    this.remoteStream = null;
    this.originalVideoTrack = null;
  }
}
