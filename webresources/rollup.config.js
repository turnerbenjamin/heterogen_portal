// rollup.config.js
import path from "path";
import { createRequire } from "module";
const require = createRequire(import.meta.url);
const glob = require("glob");
import resolve from "@rollup/plugin-node-resolve";
import commonjs from "@rollup/plugin-commonjs";
import typescript from "@rollup/plugin-typescript";

const inputFiles = glob.sync("./scripts/*.ts");

const root = path.resolve(import.meta.dirname, "..");

export default inputFiles.map((file) => ({
  input: file,
  output: {
    file: path.join(root, "cmd/static/js", file.replace(/\.ts$/, ".js")),
    format: "iife",
    name: path.basename(file, ".ts"),
    sourcemap: false,
    globals: {
      "htmx.org": "htmx",
    },
  },
  plugins: [resolve(), commonjs(), typescript({ tsconfig: "./tsconfig.json" })],
  onwarn(warning, warn) {
    const id =
      (warning && (warning.id || (warning.loc && warning.loc.file))) || "";
    // suppress warnings that originate from node_modules (dependencies)
    if (String(id).includes("node_modules")) return;
    // keep other warnings
    warn(warning);
  },
}));
