/// <reference types="@testing-library/jest-dom/vitest" />
import "@testing-library/jest-dom/vitest";
import { configure } from "@testing-library/react";

// React 19 の StrictMode は開発時にコンポーネントを複数回マウントするため
// テストのクエリが重複要素を見つけてしまう。テスト環境では無効化する。
configure({ reactStrictMode: false });
