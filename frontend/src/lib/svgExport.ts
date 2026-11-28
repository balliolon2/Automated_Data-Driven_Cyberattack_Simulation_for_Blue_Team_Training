export interface SvgExportOptions {
  width: number;
  height: number;
  title?: string;
  backgroundColor?: string;
}

const DEFAULT_SVG_BG = "#18181b";
const DEFAULT_TEXT_FILL = "#e4e4e7";
const DEFAULT_AXIS_STROKE = "#52525b";

function escapeXml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

/** Strips transient interactive tooltip nodes and normalizes text/stroke fills on a cloned SVG node. */
export function normalizeClonedSvgForExport(clone: SVGSVGElement): SVGSVGElement {
  clone
    .querySelectorAll(".recharts-tooltip-wrapper, .recharts-tooltip-cursor, .recharts-active-dot")
    .forEach((node) => node.remove());

  clone.querySelectorAll("text, tspan").forEach((node) => {
    if (!node.getAttribute("fill") && !node.getAttribute("style")?.includes("fill:")) {
      node.setAttribute("fill", DEFAULT_TEXT_FILL);
    }
  });

  clone.querySelectorAll(".recharts-cartesian-axis-line, .recharts-cartesian-axis-tick-line").forEach((node) => {
    if (!node.getAttribute("stroke")) {
      node.setAttribute("stroke", DEFAULT_AXIS_STROKE);
    }
  });

  return clone;
}

/** Enhances serialized Recharts SVG XML into a portable standalone file with backdrop and Thai font fallback. */
export function serializeSvgForExport(rawSvg: string, options: SvgExportOptions): string {
  const width = Math.max(120, Math.round(options.width || 640));
  const height = Math.max(120, Math.round(options.height || 320));
  const bg = options.backgroundColor ?? DEFAULT_SVG_BG;

  let svg = rawSvg.trim();
  const openTagMatch = svg.match(/^<svg\b([^>]*)>/i);
  if (!openTagMatch) {
    throw new Error("Expected <svg> root element for standalone export");
  }

  let attrs = openTagMatch[1] ?? "";
  if (!/\bxmlns=/.test(attrs)) {
    attrs += ' xmlns="http://www.w3.org/2000/svg"';
  }
  if (!/\bwidth=/.test(attrs)) {
    attrs += ` width="${width}"`;
  } else {
    attrs = attrs.replace(/\bwidth="[^"]*"/, `width="${width}"`);
  }
  if (!/\bheight=/.test(attrs)) {
    attrs += ` height="${height}"`;
  } else {
    attrs = attrs.replace(/\bheight="[^"]*"/, `height="${height}"`);
  }
  if (!/\bviewBox=/.test(attrs)) {
    attrs += ` viewBox="0 0 ${width} ${height}"`;
  }

  const titleNode = options.title ? `<title>${escapeXml(options.title)}</title>` : "";
  const injectedHead =
    titleNode +
    `<defs><style>text, tspan { font-family: "Geist", "Inter", "IBM Plex Sans Thai", system-ui, -apple-system, sans-serif; }</style></defs>` +
    `<rect x="0" y="0" width="100%" height="100%" fill="${bg}"/>`;

  svg = `<svg${attrs}>${injectedHead}${svg.slice(openTagMatch[0].length)}`;
  return `<?xml version="1.0" encoding="UTF-8"?>\n${svg}`;
}

/** Deep-clones the Recharts SVG inside containerEl, normalizes it, and triggers a .svg download. */
export function exportSvgElement(
  containerEl: HTMLElement | null,
  filename: string,
  title?: string,
): boolean {
  if (!containerEl || typeof window === "undefined") return false;
  const svgEl =
    containerEl.querySelector<SVGSVGElement>(".recharts-wrapper svg.recharts-surface") ??
    containerEl.querySelector<SVGSVGElement>(".recharts-wrapper svg") ??
    containerEl.querySelector<SVGSVGElement>("svg");
  if (!svgEl) return false;

  const rect = svgEl.getBoundingClientRect();
  const width = rect.width > 0 ? rect.width : Number(svgEl.getAttribute("width")) || 640;
  const height = rect.height > 0 ? rect.height : Number(svgEl.getAttribute("height")) || 320;

  const clone = normalizeClonedSvgForExport(svgEl.cloneNode(true) as SVGSVGElement);
  const rawXml = new XMLSerializer().serializeToString(clone);
  const standaloneXml = serializeSvgForExport(rawXml, {
    width,
    height,
    title,
    backgroundColor: DEFAULT_SVG_BG,
  });

  const safeName = filename.toLowerCase().endsWith(".svg") ? filename : `${filename}.svg`;
  const blob = new Blob([standaloneXml], { type: "image/svg+xml;charset=utf-8" });
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = safeName;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.URL.revokeObjectURL(url);
  return true;
}
