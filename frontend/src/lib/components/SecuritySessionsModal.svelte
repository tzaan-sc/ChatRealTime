<script>
  import { onMount } from 'svelte';
  import { apiRequest } from '../services/api';
  import { ShieldAlert, Laptop, Smartphone, Globe, LogOut, CheckCircle2, AlertTriangle, RefreshCw, X } from 'lucide-svelte';

  export let isOpen = false;
  export let onClose = () => {};

  let loading = false;
  let error = '';
  let successMsg = '';
  let devices = [];
  let alerts = [];

  async function loadSecurityData() {
    loading = true;
    error = '';
    try {
      const res = await apiRequest('/auth/security/sessions');
      devices = res.devices || [];
      alerts = res.alerts || [];
    } catch (err) {
      error = err.message || 'Không thể tải thông tin bảo mật phiên';
    } finally {
      loading = false;
    }
  }

  async function handleRevokeOthers() {
    const currentRefresh = localStorage.getItem('refresh_token') || '';
    loading = true;
    error = '';
    successMsg = '';
    try {
      const res = await apiRequest('/auth/security/revoke-others', 'POST', {
        current_refresh_token: currentRefresh
      });
      successMsg = res.message || 'Đã đăng xuất khỏi tất cả các thiết bị khác thành công!';
      await loadSecurityData();
    } catch (err) {
      error = err.message || 'Lỗi khi thu hồi phiên thiết bị khác';
    } finally {
      loading = false;
    }
  }

  $: if (isOpen) {
    loadSecurityData();
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-md animate-fadeIn">
    <!-- Modal Card -->
    <div class="relative w-full max-w-2xl bg-slate-900/95 border border-slate-700/80 rounded-2xl shadow-2xl shadow-cyan-500/10 overflow-hidden flex flex-col max-h-[90vh]">
      
      <!-- Modal Header -->
      <div class="px-6 py-5 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-gradient-to-tr from-amber-500/20 to-cyan-500/20 rounded-xl border border-amber-500/30 text-amber-400">
            <ShieldAlert size={24} />
          </div>
          <div>
            <h2 class="text-xl font-bold text-white tracking-wide flex items-center gap-2">
              Bảo Mật Phiên & Thiết Bị Đăng Nhập
              <span class="text-xs px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 font-medium">SSO & Rotation</span>
            </h2>
            <p class="text-xs text-slate-400 mt-0.5">Giám sát các phiên làm việc, cảnh báo thiết bị lạ và thu hồi quyền truy cập</p>
          </div>
        </div>
        <button 
          on:click={onClose}
          class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
          title="Đóng"
        >
          <X size={20} />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 overflow-y-auto space-y-6 custom-scrollbar">
        {#if error}
          <div class="p-3.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-300 text-sm flex items-center gap-2.5">
            <AlertTriangle size={18} class="shrink-0 text-red-400" />
            <span>{error}</span>
          </div>
        {/if}

        {#if successMsg}
          <div class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-sm flex items-center gap-2.5">
            <CheckCircle2 size={18} class="shrink-0 text-emerald-400" />
            <span>{successMsg}</span>
          </div>
        {/if}

        <!-- Section 1: Thiết bị đang đăng nhập -->
        <div>
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-semibold uppercase tracking-wider text-slate-300 flex items-center gap-2">
              <Laptop size={16} class="text-cyan-400" />
              Các phiên đang hoạt động ({devices.length})
            </h3>
            <button
              on:click={loadSecurityData}
              disabled={loading}
              class="text-xs text-slate-400 hover:text-cyan-400 flex items-center gap-1 transition-colors disabled:opacity-50"
            >
              <RefreshCw size={13} class={loading ? 'animate-spin' : ''} />
              Làm mới
            </button>
          </div>

          <div class="space-y-2.5">
            {#if devices.length === 0 && !loading}
              <div class="p-4 rounded-xl bg-slate-800/40 border border-slate-800 text-center text-slate-400 text-sm">
                Chỉ có phiên hiện tại được lưu trữ.
              </div>
            {/if}

            {#each devices as dev}
              <div class="p-4 rounded-xl bg-slate-800/50 border {dev.is_current ? 'border-cyan-500/40 bg-cyan-950/20' : 'border-slate-800'} flex items-center justify-between gap-4 transition-all">
                <div class="flex items-center gap-3.5">
                  <div class="p-2.5 rounded-lg {dev.is_current ? 'bg-cyan-500/20 text-cyan-300' : 'bg-slate-700/50 text-slate-400'}">
                    {#if dev.device_name?.toLowerCase().includes('iphone') || dev.device_name?.toLowerCase().includes('android')}
                      <Smartphone size={20} />
                    {:else}
                      <Laptop size={20} />
                    {/if}
                  </div>
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-semibold text-white">{dev.device_name || 'Thiết bị không xác định'}</span>
                      {#if dev.is_current}
                        <span class="text-[10px] uppercase font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                          Phiên hiện tại
                        </span>
                      {/if}
                    </div>
                    <div class="flex items-center gap-2 text-xs text-slate-400 mt-1">
                      <Globe size={13} class="text-slate-500" />
                      <span>{dev.location || dev.ip_address}</span>
                      <span class="text-slate-600">•</span>
                      <span>Hoạt động: {new Date(dev.last_active).toLocaleString('vi-VN')}</span>
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          </div>

          {#if devices.length > 1}
            <div class="mt-3 text-right">
              <button
                on:click={handleRevokeOthers}
                disabled={loading}
                class="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold text-amber-300 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/30 rounded-xl transition-all disabled:opacity-50"
              >
                <LogOut size={14} />
                Đăng xuất khỏi tất cả thiết bị khác
              </button>
            </div>
          {/if}
        </div>

        <!-- Section 2: Lịch sử cảnh báo an ninh -->
        <div>
          <h3 class="text-sm font-semibold uppercase tracking-wider text-slate-300 mb-3 flex items-center gap-2">
            <AlertTriangle size={16} class="text-amber-400" />
            Lịch sử cảnh báo thiết bị lạ ({alerts.length})
          </h3>

          <div class="space-y-2">
            {#if alerts.length === 0}
              <div class="p-4 rounded-xl bg-slate-800/40 border border-slate-800 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
                <CheckCircle2 size={16} class="text-emerald-400" />
                Tài khoản an toàn. Chưa ghi nhận cảnh báo đăng nhập bất thường nào.
              </div>
            {/if}

            {#each alerts as a}
              <div class="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-200 text-xs flex items-start gap-3">
                <ShieldAlert size={18} class="shrink-0 text-amber-400 mt-0.5" />
                <div class="flex-1">
                  <p class="font-medium text-amber-300">{a.message}</p>
                  <p class="text-slate-400 mt-1">
                    Thời gian: {new Date(a.created_at).toLocaleString('vi-VN')} • IP: {a.ip_address}
                  </p>
                </div>
              </div>
            {/each}
          </div>
        </div>

        <!-- Section 3: Giải thích tính năng bảo vệ -->
        <div class="p-4 rounded-xl bg-slate-800/30 border border-slate-800 text-xs text-slate-400 space-y-2">
          <p class="font-semibold text-slate-300 flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-cyan-400"></span>
            Cơ chế bảo mật nâng cao đang kích hoạt:
          </p>
          <ul class="list-disc pl-4 space-y-1 text-slate-400">
            <li><strong class="text-slate-300">Refresh Token Rotation:</strong> Mỗi phiên được cấp Refresh Token dùng 1 lần, tự động thu hồi khi phát hiện tái sử dụng.</li>
            <li><strong class="text-slate-300">Cloudflare & Rate Limiting:</strong> Tự động chặn IP gửi request quá tần suất cho phép để chống tấn công brute-force.</li>
            <li><strong class="text-slate-300">Cảnh báo thiết bị lạ:</strong> Hệ thống tự động gửi email thông báo khi phát hiện đăng nhập từ IP hoặc trình duyệt mới lạ.</li>
          </ul>
        </div>

      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-4 border-t border-slate-800 bg-slate-900/50 flex justify-end">
        <button
          on:click={onClose}
          class="px-5 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm font-semibold transition-colors"
        >
          Đóng
        </button>
      </div>

    </div>
  </div>
{/if}

<style>
  @keyframes fadeIn {
    from { opacity: 0; transform: scale(0.97); }
    to { opacity: 1; transform: scale(1); }
  }
  .animate-fadeIn {
    animation: fadeIn 0.18s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(100, 116, 139, 0.4);
    border-radius: 9999px;
  }
</style>
