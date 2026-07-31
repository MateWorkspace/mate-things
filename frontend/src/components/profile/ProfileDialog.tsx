"use client";

import { useCallback, useEffect, useState } from "react";

import LogoutButton from "@/components/layout/LogoutButton";
import Button from "@/components/ui/button";
import Dialog from "@/components/ui/dialog";
import Tabs, { getTabId, type Tab } from "@/components/ui/tabs";
import type { UserResponse } from "@/lib/api/users";

import ProfileForm from "./ProfileForm";
import ProfileView from "./ProfileView";
import SecurityForm from "./SecurityForm";

interface ProfileDialogProps {
  open: boolean;
  user: UserResponse;
  permissions: readonly string[];
  roleName?: string;
  onClose?: () => void;
}

const PROFILE_TAB: Tab = {
  id: "profile",
  label: "Profile",
  panelId: "profile-panel",
};

const SECURITY_TAB: Tab = {
  id: "security",
  label: "Security",
  panelId: "security-panel",
};

export default function ProfileDialog({
  open,
  user,
  permissions,
  roleName,
  onClose = () => undefined,
}: ProfileDialogProps) {
  const canEditProfile = permissions.includes("profile:set");
  const canChangePassword = permissions.includes("profile_security:set");
  const tabs = canChangePassword ? [PROFILE_TAB, SECURITY_TAB] : [PROFILE_TAB];
  const [activeTab, setActiveTab] = useState(PROFILE_TAB.id);
  const [editing, setEditing] = useState(false);

  useEffect(() => {
    if (!canEditProfile) {
      setEditing(false);
    }
  }, [canEditProfile]);

  useEffect(() => {
    if (!canChangePassword && activeTab === SECURITY_TAB.id) {
      setActiveTab(PROFILE_TAB.id);
    }
  }, [activeTab, canChangePassword]);

  const closeDialog = useCallback(() => {
    setActiveTab(PROFILE_TAB.id);
    setEditing(false);
    onClose();
  }, [onClose]);

  return (
    <Dialog
      open={open}
      onClose={closeDialog}
      title="Your profile"
      variant="sheet"
    >
      <Tabs
        tabs={tabs}
        activeTab={activeTab}
        onChange={(tab) => {
          setActiveTab(tab);
          setEditing(false);
        }}
        ariaLabel="Account settings"
      />

      {activeTab === PROFILE_TAB.id ? (
        <section
          id={PROFILE_TAB.panelId}
          role="tabpanel"
          aria-labelledby={getTabId(PROFILE_TAB.panelId)}
          className="pt-5"
        >
          {editing ? (
            <ProfileForm user={user} onCancel={() => setEditing(false)} />
          ) : (
            <ProfileView
              user={user}
              roleName={roleName}
              canEdit={canEditProfile}
              onEdit={() => setEditing(true)}
            />
          )}
        </section>
      ) : (
        <section
          id={SECURITY_TAB.panelId}
          role="tabpanel"
          aria-labelledby={getTabId(SECURITY_TAB.panelId)}
          className="pt-5"
        >
          <SecurityForm />
        </section>
      )}

      <footer className="border-border mt-6 flex items-center justify-between gap-3 border-t pt-4">
        <LogoutButton />
        <Button type="button" variant="secondary" onClick={closeDialog}>
          Close
        </Button>
      </footer>
    </Dialog>
  );
}
