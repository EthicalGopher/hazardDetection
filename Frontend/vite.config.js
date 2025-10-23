import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      './hazard_pb.js': './hazard_pb.js?commonjs-proxy',
      './hazard_grpc_web_pb.js': './hazard_grpc_web_pb.js?commonjs-proxy',
    },
  },
})