<template>
  <div class="background-animation">
    <div class="blob blob-1"></div>
    <div class="blob blob-2"></div>
    <div class="blob blob-3"></div>
    <div class="particles">
      <div v-for="i in 12" :key="i" class="particle" :style="getRandomStyle(i)"></div>
    </div>
    <div class="grid-overlay"></div>
  </div>
</template>

<script setup>
const getRandomStyle = (i) => {
  const size = Math.random() * 4 + 2
  const duration = Math.random() * 15 + 10
  const delay = Math.random() * 5
  const left = Math.random() * 100
  const top = Math.random() * 100
  
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${left}%`,
    top: `${top}%`,
    animationDuration: `${duration}s`,
    animationDelay: `-${delay}s`,
    opacity: Math.random() * 0.5 + 0.2
  }
}
</script>

<style scoped>
.background-animation {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  pointer-events: none;
  z-index: 0;
}

.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.3;
  transition: all 0.5s ease;
}

.blob-1 {
  width: 800px;
  height: 800px;
  background: var(--brand-500);
  top: -200px;
  left: -200px;
  animation: move1 20s infinite alternate;
}

.blob-2 {
  width: 600px;
  height: 600px;
  background: #a855f7;
  bottom: -100px;
  right: -100px;
  animation: move2 25s infinite alternate;
}

.blob-3 {
  width: 500px;
  height: 500px;
  background: #0ea5e9;
  top: 30%;
  left: 20%;
  animation: move3 18s infinite alternate;
}

.particles {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.particle {
  position: absolute;
  background: var(--brand-400, #60a5fa);
  border-radius: 50%;
  animation: float-particle infinite ease-in-out;
}

@keyframes float-particle {
  0%, 100% { transform: translate(0, 0); }
  33% { transform: translate(30px, 50px); }
  66% { transform: translate(-20px, 80px); }
}

.grid-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: 
    linear-gradient(var(--border-1) 1.5px, transparent 1px),
    linear-gradient(90deg, var(--border-1) 1.5px, transparent 1px);
  background-size: 50px 50px;
  mask-image: radial-gradient(circle at center, black, transparent 90%);
  opacity: 0.3;
}

@keyframes move1 {
  from { transform: translate(0, 0) scale(1); }
  to { transform: translate(100px, 100px) scale(1.2); }
}

@keyframes move2 {
  from { transform: translate(0, 0) scale(1.1); }
  to { transform: translate(-150px, -50px) scale(0.9); }
}

@keyframes move3 {
  from { transform: translate(0, 0); }
  to { transform: translate(200px, -100px); }
}

:global(.dark) .blob {
  opacity: 0.15;
  filter: blur(120px);
}

:global(.dark) .particle {
  background: var(--brand-400);
  opacity: 0.15 !important;
}
</style>
