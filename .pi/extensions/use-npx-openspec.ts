import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import hookModule from "../../.agents/hooks/use_npx_openspec.js";

const { createPiExtension } = hookModule as {
  createPiExtension: (pi: ExtensionAPI) => void;
};

export default function (pi: ExtensionAPI) {
  createPiExtension(pi);
}
