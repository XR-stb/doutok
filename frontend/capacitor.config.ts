import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "com.doutok.app",
  appName: "DouTok",
  webDir: "dist",
  server: {
    androidScheme: "https",
    // Debug 模式下连本地开发服务器
    ...(process.env.NODE_ENV === "development" && {
      url: "http://10.0.2.2:3000",
      cleartext: true,
    }),
  },
  plugins: {
    SplashScreen: {
      launchAutoHide: true,
      launchShowDuration: 2000,
      backgroundColor: "#000000",
    },
    StatusBar: {
      style: "DARK",
      backgroundColor: "#000000",
    },
  },
};

export default config;
