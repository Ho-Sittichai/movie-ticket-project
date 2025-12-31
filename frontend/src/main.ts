import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Toast, { type PluginOptions, POSITION } from "vue-toastification"
import "vue-toastification/dist/index.css"
import './style.css'
import App from './App.vue'
import router from './router'

const options: PluginOptions = {
    position: POSITION.BOTTOM_LEFT,
    timeout: 3000,
    closeOnClick: true,
    pauseOnFocusLoss: true,
    pauseOnHover: true,
    draggable: true,
    draggablePercent: 0.6,
    showCloseButtonOnHover: false,
    hideProgressBar: false,
    closeButton: "button",
    icon: true,
    rtl: false
}

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Toast, options)

app.mount('#app')
