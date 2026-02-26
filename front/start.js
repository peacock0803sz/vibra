#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { createRequire } from "node:module";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);
const serverBundle = resolve(__dirname, "build", "server", "index.js");
const serveBin = require.resolve("@react-router/serve/bin");

execFileSync(process.execPath, [serveBin, serverBundle], {
  stdio: "inherit",
  env: process.env,
});
