import "@testing-library/jest-dom/vitest";
import { configure } from "@testing-library/react";

// React 19's StrictMode mounts components multiple times during development,
// causing test queries to find duplicate elements. Disable it in the test environment.
configure({ reactStrictMode: false });
