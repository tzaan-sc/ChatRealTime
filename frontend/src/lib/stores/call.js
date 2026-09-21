import { writable } from 'svelte/store';
import { WebRTCManager } from '../utils/webrtc';
import { wsService } from '../services/websocket';
import { startRingtone, stopRingtone, startCallingTone, stopCallingTone } from '../utils/sound';

export const callState = writable('idle'); // 'idle', 'outgoing', 'incoming', 'connected'
export const callType = writable('video'); // 'audio', 'video'
export const callPartner = writable(null); // { id, name, avatar }
export const localStream = writable(null);
export const remoteStream = writable(null);
export const isMuted = writable(false);
export const isCameraOff = writable(false);
export const isScreenSharing = writable(false);
export const callDuration = writable(0);

let webrtcManager = null;
let durationTimer = null;

// Khởi tạo cuộc gọi đi (Outgoing Call)
export async function startCall(partner, type = 'video') {
  if (!partner || !partner.id) return;

  endCallCleanup();

  callPartner.set(partner);
  callType.set(type);
  callState.set('outgoing');
  startCallingTone();

  webrtcManager = new WebRTCManager({
    onRemoteStream: (stream) => {
      remoteStream.set(stream);
      callConnected();
    },
    onIceCandidate: (candidate) => {
      wsService.send('call:ice_candidate', {
        receiver_id: partner.id,
        candidate: candidate
      });
    },
    onConnectionStateChange: (state) => {
      if (state === 'connected') {
        callConnected();
      } else if (state === 'disconnected' || state === 'failed' || state === 'closed') {
        endCall();
      }
    }
  });

  try {
    const stream = await webrtcManager.initLocalStream(type);
    localStream.set(stream);

    // Gửi yêu cầu gọi điện qua WebSocket
    wsService.send('call:request', {
      receiver_id: partner.id,
      call_type: type
    });
  } catch (err) {
    console.error('Không thể bắt đầu cuộc gọi:', err);
    alert('Không thể truy cập camera hoặc microphone: ' + err.message);
    endCall();
  }
}

// Khi có cuộc gọi đến (Incoming Call)
export function handleIncomingCall(payload) {
  if (!payload || !payload.sender_id) return;

  // Nếu đang trong cuộc gọi khác thì tự động từ chối báo bận
  let currentState = 'idle';
  callState.subscribe((v) => (currentState = v))();
  if (currentState !== 'idle') {
    wsService.send('call:reject', {
      receiver_id: payload.sender_id,
      reason: 'busy'
    });
    return;
  }

  callPartner.set({
    id: payload.sender_id,
    name: payload.sender_name || 'Người dùng',
    avatar: payload.sender_avatar || ''
  });
  callType.set(payload.call_type || 'video');
  callState.set('incoming');
  startRingtone();
}

// Chấp nhận cuộc gọi đến
export async function acceptIncomingCall() {
  stopRingtone();
  let partner = null;
  let type = 'video';
  callPartner.subscribe((v) => (partner = v))();
  callType.subscribe((v) => (type = v))();

  if (!partner) return;

  webrtcManager = new WebRTCManager({
    onRemoteStream: (stream) => {
      remoteStream.set(stream);
      callConnected();
    },
    onIceCandidate: (candidate) => {
      wsService.send('call:ice_candidate', {
        receiver_id: partner.id,
        candidate: candidate
      });
    },
    onConnectionStateChange: (state) => {
      if (state === 'connected') {
        callConnected();
      } else if (state === 'disconnected' || state === 'failed' || state === 'closed') {
        endCall();
      }
    }
  });

  try {
    const stream = await webrtcManager.initLocalStream(type);
    localStream.set(stream);

    // Báo cho phía gọi biết mình đã chấp nhận nghe máy
    wsService.send('call:accept', {
      receiver_id: partner.id
    });
  } catch (err) {
    console.error('Lỗi khi chấp nhận cuộc gọi:', err);
    alert('Không thể truy cập camera hoặc micro: ' + err.message);
    rejectIncomingCall('permission_denied');
  }
}

// Từ chối cuộc gọi đến
export function rejectIncomingCall(reason = 'declined') {
  let partner = null;
  callPartner.subscribe((v) => (partner = v))();
  if (partner) {
    wsService.send('call:reject', {
      receiver_id: partner.id,
      reason: reason
    });
  }
  endCallCleanup();
}

// Khi phía đối phương chấp nhận nghe máy -> Caller tạo Offer
export async function handleCallAccepted() {
  stopCallingTone();
  let partner = null;
  callPartner.subscribe((v) => (partner = v))();
  if (!webrtcManager || !partner) return;

  try {
    const offer = await webrtcManager.createOffer();
    wsService.send('call:offer', {
      receiver_id: partner.id,
      sdp: offer
    });
  } catch (err) {
    console.error('Lỗi tạo offer:', err);
    endCall();
  }
}

// Khi Callee nhận được Offer từ Caller -> Callee tạo Answer
export async function handleReceiveOffer(payload) {
  if (!webrtcManager || !payload || !payload.sdp) return;
  try {
    const answer = await webrtcManager.handleOfferAndCreateAnswer(payload.sdp);
    wsService.send('call:answer', {
      receiver_id: payload.sender_id,
      sdp: answer
    });
  } catch (err) {
    console.error('Lỗi xử lý offer và tạo answer:', err);
    endCall();
  }
}

// Khi Caller nhận Answer từ Callee
export async function handleReceiveAnswer(payload) {
  if (!webrtcManager || !payload || !payload.sdp) return;
  try {
    await webrtcManager.handleAnswer(payload.sdp);
  } catch (err) {
    console.error('Lỗi xử lý answer:', err);
  }
}

// Khi nhận được ICE Candidate
export async function handleReceiveIceCandidate(payload) {
  if (webrtcManager && payload && payload.candidate) {
    await webrtcManager.addIceCandidate(payload.candidate);
  }
}

// Khi cuộc gọi kết nối thành công
function callConnected() {
  stopCallingTone();
  stopRingtone();
  callState.set('connected');

  if (!durationTimer) {
    callDuration.set(0);
    durationTimer = setInterval(() => {
      callDuration.update((d) => d + 1);
    }, 1000);
  }
}

// Bật/Tắt Micro
export function toggleMute() {
  isMuted.update((muted) => {
    const next = !muted;
    if (webrtcManager) {
      webrtcManager.toggleAudio(!next);
    }
    return next;
  });
}

// Bật/Tắt Camera
export function toggleCamera() {
  isCameraOff.update((off) => {
    const next = !off;
    if (webrtcManager) {
      webrtcManager.toggleVideo(!next);
    }
    return next;
  });
}

// Chia sẻ màn hình
export async function toggleScreenShare() {
  let sharing = false;
  isScreenSharing.subscribe((v) => (sharing = v))();

  if (sharing) {
    if (webrtcManager) await webrtcManager.stopScreenShare();
    isScreenSharing.set(false);
  } else {
    if (webrtcManager) {
      const stream = await webrtcManager.startScreenShare();
      if (stream) {
        isScreenSharing.set(true);
      }
    }
  }
}

// Kết thúc cuộc gọi từ phía mình
export function endCall() {
  let partner = null;
  callPartner.subscribe((v) => (partner = v))();
  if (partner) {
    wsService.send('call:hangup', {
      receiver_id: partner.id
    });
  }
  endCallCleanup();
}

// Khi đối phương cúp máy hoặc từ chối
export function handleRemoteHangup(payload) {
  endCallCleanup();
}

// Dọn dẹp trạng thái
function endCallCleanup() {
  stopCallingTone();
  stopRingtone();

  if (durationTimer) {
    clearInterval(durationTimer);
    durationTimer = null;
  }

  if (webrtcManager) {
    webrtcManager.close();
    webrtcManager = null;
  }

  callState.set('idle');
  callPartner.set(null);
  localStream.set(null);
  remoteStream.set(null);
  isMuted.set(false);
  isCameraOff.set(false);
  isScreenSharing.set(false);
  callDuration.set(0);
}
