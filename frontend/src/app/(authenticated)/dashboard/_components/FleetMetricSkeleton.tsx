import Card from "@/components/ui/card";
import { LoadingState } from "@/components/ui/states";

interface FleetMetricSkeletonProps {
  title: string;
}

export default function FleetMetricSkeleton({
  title,
}: FleetMetricSkeletonProps) {
  return (
    <Card aria-label={`${title} metric card`}>
      <LoadingState title={`${title} unavailable`} />
    </Card>
  );
}
