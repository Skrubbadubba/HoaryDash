import icons from "@mdi/js";
import { writeFileSync } from "fs";

const out = {};
for (const [key, path] of Object.entries(icons)) {
  if (!key.startsWith("mdi")) continue;
  const name = key
    .replace(/^mdi/, "")
    .replace(/([A-Z])/g, (m, c, i) => i === 0 ? c.toLowerCase() : "-" + c.toLowerCase());
  out[name] = path;
}
writeFileSync("mdi.json", JSON.stringify(out));