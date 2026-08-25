import type { DecodeResult } from "../lib/api";

interface Props {
  result: DecodeResult;
  index: number;
}

export function TechReadout({ result, index }: Props) {
  const typeLabel = result.content_type?.toUpperCase() ?? "UNKNOWN";
  const eccLabel =
    result.error_correction_level === "L"
      ? "REED_SOLOMON_7"
      : result.error_correction_level === "M"
        ? "REED_SOLOMON_15"
        : result.error_correction_level === "Q"
          ? "REED_SOLOMON_25"
          : "REED_SOLOMON_30";

  const bb = result.bounding_box;
  const br = bb?.bottom_right ?? [0, 0];
  const hasBounds =
    bb &&
    (bb.top_left[0] !== 0 || bb.top_left[1] !== 0 || br[0] !== 0 || br[1] !== 0);

  return (
    <div className="tech-readout">
      <div className="tech-readout-header">
        <span className="tech-readout-label">TECH_READOUT</span>
        <span className="tech-readout-id">#{index + 1}</span>
      </div>
      <div className="tech-readout-grid">
        <div className="tech-readout-row">
          <span className="tech-readout-key">TYPE_CLASSIFICATION</span>
          <span className="tech-readout-value">{typeLabel}</span>
        </div>
        <div className="tech-readout-row">
          <span className="tech-readout-key">RECOVERY_LEVEL</span>
          <span className="tech-readout-value">{eccLabel}</span>
        </div>
        {hasBounds && (
          <>
            <div className="tech-readout-row">
              <span className="tech-readout-key">OPTICAL_BOUNDS</span>
              <span className="tech-readout-value">
                [{bb.top_left[0].toFixed(0)},{bb.top_left[1].toFixed(0)}] → [{br[0].toFixed(0)},{br[1].toFixed(0)}]
              </span>
            </div>
            <div className="tech-readout-row">
              <span className="tech-readout-key">MODULE_COUNT</span>
              <span className="tech-readout-value">
                {Math.round(Math.sqrt(Math.pow(bb.top_right[0] - bb.top_left[0], 2) + Math.pow(bb.top_right[1] - bb.top_left[1], 2)))} modules
              </span>
            </div>
          </>
        )}
        <div className="tech-readout-row">
          <span className="tech-readout-key">RAW_STREAM_PAYLOAD</span>
          <span className="tech-readout-value tech-readout-payload">
            {result.content?.length > 80
              ? result.content.slice(0, 80) + "..."
              : result.content}
          </span>
        </div>
      </div>
    </div>
  );
}
