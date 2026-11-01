import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'
import './judo-theme.css'
import './days-panel.css'

const app = createApp(App)
app.use(createPinia())
app.mount('#app')