import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Users, User } from "lucide-react";
import type { Activity } from "@/types/activity";

interface ActivityParticipantsProps {
  activity: Activity;
}

export default function ActivityParticipants({
  activity,
}: ActivityParticipantsProps) {
  const participants = activity.participants || [];

  if (participants.length === 0) {
    return (
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            参与者列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            暂无参与者
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="rounded-xl shadow-lg">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Users className="h-5 w-5" />
          参与者列表 ({participants.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {participants.map((participant, index) => (
            <div
              key={participant.user_id}
              className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                  <User className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <div className="font-medium">
                    {participant.user_info?.name ||
                      `用户 ${participant.user_id}`}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {participant.user_info?.student_id || participant.user_id}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-bold text-primary">
                  {participant.credits} 学分
                </div>
                <div className="text-xs text-muted-foreground">
                  {new Date(participant.joined_at).toLocaleDateString()}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
