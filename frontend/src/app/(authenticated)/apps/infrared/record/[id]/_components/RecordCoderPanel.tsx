import type { InfraredStateCoderResponse } from "@/lib/api/infrared";

import CodeBlock from "./CodeBlock";
import DeleteStateCoderDialog from "./DeleteStateCoderDialog";

export default function RecordCoderPanel({
  coder,
  canDelete,
}: {
  coder: InfraredStateCoderResponse;
  canDelete: boolean;
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-display text-lg">{coder.summary_readme}</h3>
          <p className="text-muted-foreground text-sm">{coder.detail_readme}</p>
        </div>
        {canDelete ? <DeleteStateCoderDialog coderId={coder.id} /> : null}
      </div>
      <div>
        <h4 className="mb-2 text-sm font-semibold">Encoder</h4>
        <CodeBlock value={coder.encoder_source} label="encoder source" />
      </div>
      <div>
        <h4 className="mb-2 text-sm font-semibold">Decoder</h4>
        <CodeBlock value={coder.decoder_source} label="decoder source" />
      </div>
    </div>
  );
}
