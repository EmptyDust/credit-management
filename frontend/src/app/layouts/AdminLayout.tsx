import RoleBasedRoute from "@/components/RoleBasedRoute";
import DevToolsLayout from "@/pages/devtools";

export default function AdminLayout() {
  return (
    <RoleBasedRoute allowedRoles={["admin"]}>
      <DevToolsLayout />
    </RoleBasedRoute>
  );
}
