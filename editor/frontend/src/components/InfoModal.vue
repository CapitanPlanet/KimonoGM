<template>
  <div v-if="isOpen && !isDismissed" class="modal-overlay" @click.self="closeModal">
    <div class="info-modal-box">
      <div class="modal-title">INFORMACJA</div>
      
      <div class="modal-content">
        <!-- Możesz przekazać treść przez prop message lub slot -->
        <slot>
          <p>{{ message }}</p>
        </slot>
      </div>
      
      <div class="modal-footer">
        <label class="checkbox-label">
          <input type="checkbox" v-model="dontShowAgain" />
          <span>Więcej nie otwieraj</span>
        </label>
        <button class="kimono-btn" @click="closeModal">OK</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    isOpen: {
      type: Boolean,
      required: true
    },
    storageKey: {
      type: String,
      required: true // Unikalny klucz np. "launcher_welcome_seen"
    },
    message: {
      type: String,
      default: 'Witaj w Kimono Game Maker!'
    }
  },
  data() {
    return {
      dontShowAgain: false,
      isDismissed: false
    }
  },
  created() {
    // Sprawdzamy czy użytkownik zaznaczył wcześniej "więcej nie otwieraj"
    if (localStorage.getItem(this.storageKey) === 'true') {
      this.isDismissed = true
    }
  },
  methods: {
    closeModal() {
      if (this.dontShowAgain) {
        localStorage.setItem(this.storageKey, 'true')
      }
      this.isDismissed = true
      this.$emit('close')
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.info-modal-box {
  background: #0d0d0d;
  border: 2px solid #FFEB3B;
  box-shadow: 6px 6px 0 #FF2D7B;
  width: 580px;
  max-width: 90vw;
  padding: 24px;
  font-family: 'Courier New', monospace;
  color: #fff;
}
.modal-title {
  font-size: 13px;
  font-weight: 900;
  color: #FFEB3B;
  letter-spacing: 2px;
  margin-bottom: 16px;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(255, 235, 59, 0.2);
  padding-bottom: 8px;
}
.modal-content {
  font-size: 13px;
  line-height: 1.6;
  color: #ddd;
  margin-bottom: 24px;
  max-height: 50vh;
  overflow-y: auto;
}
.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 16px;
}
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #aaa;
  cursor: pointer;
  user-select: none;
}
.checkbox-label input {
  accent-color: #FFEB3B;
  width: 16px;
  height: 16px;
  cursor: pointer;
}
.kimono-btn {
  background: #FFEB3B;
  color: black;
  font-family: 'Courier New', monospace;
  font-weight: 900;
  font-size: 13px;
  padding: 10px 24px;
  border: 2px solid #FF2D7B;
  box-shadow: 3px 3px 0 #00E5FF;
  cursor: pointer;
}
.kimono-btn:hover {
  transform: translate(-2px, -2px);
  box-shadow: 5px 5px 0 #00E5FF;
  background: #FFF176;
}
</style>