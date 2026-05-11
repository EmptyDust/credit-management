import RoleBasedRoute from "@/components/RoleBasedRoute";
import DevtoolsLayout from "@/modules/devtools/pages/DevtoolsLayout";

export default function AdminLayout() {
  return (
    <RoleBasedRoute allowedRoles={["admin"]}>
      <DevtoolsLayout />
    </RoleBasedRoute>
  );
}
