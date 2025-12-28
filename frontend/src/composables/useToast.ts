import { ref } from 'vue';

const isVisible = ref(false);
const message = ref('');
const type = ref<'info' | 'error' | 'success'>('info');
let timeoutId: number | undefined;

export function useToast() {
  const showToast = (msg: string, toastType: 'info' | 'error' | 'success' = 'info', duration = 4000) => {
    message.value = msg;
    type.value = toastType;
    isVisible.value = true;

    if (timeoutId) clearTimeout(timeoutId);

    timeoutId = setTimeout(() => {
      isVisible.value = false;
    }, duration);
  };

  const hideToast = () => {
    isVisible.value = false;
    if (timeoutId) clearTimeout(timeoutId);
  };

  return {
    isVisible,
    message,
    type,
    showToast,
    hideToast,
  };
}
