import tailwindcss from "@tailwindcss/vite";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()],
  resolve: {
    conditions: ["browser"]
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.{test,spec}.{ts,js}"],
    setupFiles: ["./vitest.setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/lib/api/**/*.ts", "src/lib/auth/**/*.ts", "src/lib/navigation.ts", "src/lib/utils.ts"],
      exclude: ["**/*.test.ts", "**/*.spec.ts"],
      thresholds: {
        statements: 80,
        lines: 80,
        functions: 80,
        branches: 70
      }
    }
  }
});
