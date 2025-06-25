# 测试参与者删除功能
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTA5NTUwODgsImlhdCI6MTc1MDg2ODY4OCwidXNlcl9pZCI6ImEzZTM4M2MzLWE5NGQtNDJmNi1hMGMwLWNmMzRlNmEwY2Y3NyIsInVzZXJfdHlwZSI6InN0dWRlbnQiLCJ1c2VybmFtZSI6InN0dWRlbnQwMDIifQ.JHXi91SISvdQMgL7_zACfu9-NkFo-oahfQvGkEuHCHo"
$activityId = "b04dcf08-3286-4c4f-be5c-8d5c3b7394e9"
$participantUserId = "e8408b01-bc68-4f54-9f36-d5b22bbdb326"

Write-Host "1. 查看当前参与者列表..."
$participantsResponse = curl -s -H "Authorization: Bearer $token" "http://localhost:8083/api/activities/$activityId/participants"
Write-Host "参与者列表: $participantsResponse"

Write-Host "`n2. 测试删除参与者..."
$removeResponse = curl -s -X DELETE -H "Authorization: Bearer $token" "http://localhost:8083/api/activities/$activityId/participants/$participantUserId"
Write-Host "删除响应: $removeResponse"

Write-Host "`n3. 再次查看参与者列表..."
$participantsResponse2 = curl -s -H "Authorization: Bearer $token" "http://localhost:8083/api/activities/$activityId/participants"
Write-Host "删除后的参与者列表: $participantsResponse2"

Write-Host "`n4. 重新添加参与者..."
$participantData = @{
    user_ids = @($participantUserId)
    credits  = 2.0
} | ConvertTo-Json

$addResponse = curl -s -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d $participantData "http://localhost:8083/api/activities/$activityId/participants"
Write-Host "重新添加响应: $addResponse" 