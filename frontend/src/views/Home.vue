<template>
  <div class="home-layout">
    <el-container>
      <el-header>
        <el-text class="home-header-text" size="large">{{ $t('home.greeting') }}</el-text>
      </el-header>
      <el-main>{{ $t('home.main') }}</el-main>
    </el-container>
  </div>
</template>

<style scoped>
    .home-layout {
        margin: 10px;
    }
</style>
