'use client';

import { useState, useEffect, useCallback } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';
import {
  UserGroupIcon,
  PlusIcon,
  EnvelopeIcon,
  TrashIcon,
  XMarkIcon,
  LockClosedIcon,
  SparklesIcon,
  ClockIcon,
  LinkIcon,
} from '@heroicons/react/24/outline';
import { PLAN_NAMES } from '@/lib/plan-utils';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface RoleObj {
  id: number;
  name: string;
  display_name: string;
}

interface TeamMember {
  id: string;         // user ID (int64 serialised as string)
  name: string;
  email: string;
  role: RoleObj | string;
  status: string;
  joined_at: string | null;
}

interface PendingInvitation {
  id: number;
  email: string;
  role: RoleObj | string;
  status: string;
  expires_at: string;
}

interface Team {
  id: number;
  name: string;
  display_name?: string;
  description?: string;
  owner_id: string; // int64 as string
  created_at: string;
}

interface TeamLink {
  id: number;
  short_code: string;
  short_url: string;
  original_url: string;
  click_count: number;
  created_by?: number;
  created_at: string;
}

function roleLabel(role: RoleObj | string | undefined): string {
  if (!role) return 'Member';
  if (typeof role === 'string') return role;
  return role.display_name || role.name || 'Member';
}

export default function TeamsPage() {
  const { user, getAccessToken } = useAuth();
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [pendingInvites, setPendingInvites] = useState<PendingInvitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [teamName, setTeamName] = useState('');
  const [teamDesc, setTeamDesc] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('member');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [teamLinks, setTeamLinks] = useState<TeamLink[]>([]);
  const [linksLoading, setLinksLoading] = useState(false);

  const isPro = user?.account_type === 'premium' || user?.account_type === 'enterprise';

  const authHeaders = useCallback(
    () => ({ Authorization: `Bearer ${getAccessToken()}`, 'Content-Type': 'application/json' }),
    [getAccessToken]
  );

  const fetchTeams = useCallback(async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/teams`, { headers: authHeaders() });
      if (res.ok) {
        const data = await res.json();
        const list: Team[] = data.teams ?? [];
        setTeams(list);
        if (list.length > 0 && !selectedTeam) setSelectedTeam(list[0]);
      }
    } finally {
      setLoading(false);
    }
  }, [authHeaders, selectedTeam]);

  useEffect(() => {
    if (user) fetchTeams();
  }, [user, fetchTeams]);

  const fetchMembers = useCallback(async (teamId: number) => {
    const res = await fetch(`${API_BASE_URL}/api/v1/teams/${teamId}`, { headers: authHeaders() });
    if (res.ok) {
      const data = await res.json();
      // API returns members at top-level AND nested in team
      setMembers(data.members ?? data.team?.members ?? []);
      setPendingInvites(data.pending_invitations ?? []);
    }
  }, [authHeaders]);

  useEffect(() => {
    if (selectedTeam) fetchMembers(selectedTeam.id);
  }, [selectedTeam, fetchMembers]);

  const fetchTeamLinks = useCallback(async (teamId: number) => {
    setLinksLoading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/teams/${teamId}/links`, { headers: authHeaders() });
      if (res.ok) {
        const data = await res.json();
        setTeamLinks(data.links ?? []);
      } else {
        setTeamLinks([]);
      }
    } finally {
      setLinksLoading(false);
    }
  }, [authHeaders]);

  useEffect(() => {
    if (selectedTeam) fetchTeamLinks(selectedTeam.id);
  }, [selectedTeam, fetchTeamLinks]);

  const handleCreateTeam = async () => {
    if (!teamName.trim()) return;
    setSubmitting(true);
    setError('');
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/teams`, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({ name: teamName.trim(), display_name: teamName.trim(), description: teamDesc.trim() }),
      });
      const data = await res.json();
      if (!res.ok) { setError(data.message ?? 'Failed to create team'); return; }
      setShowCreateModal(false);
      setTeamName(''); setTeamDesc('');
      await fetchTeams();
    } finally {
      setSubmitting(false);
    }
  };

  const handleInvite = async () => {
    if (!inviteEmail.trim() || !selectedTeam) return;
    setSubmitting(true);
    setError('');
    try {
      const roleIdMap: Record<string, number> = { admin: 2, member: 3, viewer: 4 };
      const res = await fetch(`${API_BASE_URL}/api/v1/teams/${selectedTeam.id}/invite`, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({
          email: inviteEmail.trim(),
          role: inviteRole,
          role_id: roleIdMap[inviteRole] ?? 3,
        }),
      });
      const data = await res.json();
      if (!res.ok) { setError(data.message ?? 'Failed to send invitation'); return; }
      setShowInviteModal(false);
      setInviteEmail('');
      await fetchMembers(selectedTeam.id);
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemoveMember = async (teamId: number, userId: string) => {
    if (!confirm('Remove this member from the team?')) return;
    await fetch(`${API_BASE_URL}/api/v1/teams/${teamId}/members/${userId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    await fetchMembers(teamId);
  };

  const handleCancelInvitation = async (teamId: number, invitationId: number) => {
    if (!confirm('Cancel this invitation?')) return;
    await fetch(`${API_BASE_URL}/api/v1/teams/${teamId}/invitations/${invitationId}`, {
      method: 'DELETE',
      headers: authHeaders(),
    });
    await fetchMembers(teamId);
  };

  if (!user) return null;

  if (!isPro) {
    return (
      <Layout>
        <div className="min-h-screen bg-background py-8">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center space-x-3 mb-8">
              <div className="p-2 bg-primary/10 rounded-lg">
                <UserGroupIcon className="h-6 w-6 text-primary" />
              </div>
              <h1 className="text-2xl font-bold text-foreground">Team Collaboration</h1>
            </div>
            <div className="bg-card border border-border rounded-xl p-12 text-center">
              <div className="mx-auto w-14 h-14 bg-primary/10 rounded-full flex items-center justify-center mb-4">
                <LockClosedIcon className="h-7 w-7 text-primary" />
              </div>
              <h2 className="text-xl font-semibold text-foreground mb-2">{PLAN_NAMES.premium} Feature</h2>
              <p className="text-muted-foreground mb-2 max-w-md mx-auto">
                Invite team members, assign roles, and collaborate on links together.
              </p>
              <p className="text-sm text-muted-foreground mb-6">
                Upgrade to <strong>{PLAN_NAMES.premium}</strong> or <strong>{PLAN_NAMES.enterprise}</strong> to unlock teams.
              </p>
              <Link
                href="/billing"
                className="inline-flex items-center space-x-2 px-6 py-2.5 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 font-medium transition-colors"
              >
                <SparklesIcon className="h-4 w-4" />
                <span>Upgrade Plan</span>
              </Link>
            </div>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">

          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <UserGroupIcon className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-foreground">Teams</h1>
                <p className="text-sm text-muted-foreground">Collaborate with your team members</p>
              </div>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center space-x-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
            >
              <PlusIcon className="h-4 w-4" />
              <span>New Team</span>
            </button>
          </div>

          {loading ? (
            <div className="space-y-3">
              {[1, 2].map(i => <div key={i} className="h-20 bg-card border border-border rounded-lg animate-pulse" />)}
            </div>
          ) : teams.length === 0 ? (
            <div className="bg-card border border-border rounded-xl p-12 text-center">
              <UserGroupIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-1">No teams yet</h3>
              <p className="text-sm text-muted-foreground mb-4">Create a team to collaborate with others.</p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
              >
                Create Your First Team
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Team list */}
              <div className="space-y-2">
                {teams.map(t => (
                  <button
                    key={t.id}
                    onClick={() => setSelectedTeam(t)}
                    className={`w-full text-left p-4 rounded-lg border transition-colors ${
                      selectedTeam?.id === t.id
                        ? 'bg-primary/10 border-primary/30'
                        : 'bg-card border-border hover:bg-muted'
                    }`}
                  >
                    <p className="font-medium text-foreground">{t.display_name || t.name}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {members.length + pendingInvites.length} members
                    </p>
                  </button>
                ))}
              </div>

              {/* Team detail */}
              {selectedTeam && (
                <div className="lg:col-span-2 bg-card border border-border rounded-lg p-5">
                  <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold text-foreground">
                      {selectedTeam.display_name || selectedTeam.name}
                    </h2>
                    <button
                      onClick={() => setShowInviteModal(true)}
                      className="flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                    >
                      <EnvelopeIcon className="h-3.5 w-3.5" />
                      <span>Invite</span>
                    </button>
                  </div>

                  {/* Active members */}
                  <div className="space-y-1 mb-4">
                    {members.length === 0 && pendingInvites.length === 0 ? (
                      <p className="text-sm text-muted-foreground">No members yet. Send an invitation.</p>
                    ) : members.map(m => (
                      <div key={m.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                        <div>
                          <p className="text-sm font-medium text-foreground">{m.name || m.email}</p>
                          <p className="text-xs text-muted-foreground">{m.email} · {roleLabel(m.role)}</p>
                        </div>
                        {m.id !== selectedTeam.owner_id && (
                          <button
                            onClick={() => handleRemoveMember(selectedTeam.id, m.id)}
                            className="p-1.5 text-destructive hover:bg-destructive/10 rounded transition-colors"
                          >
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    ))}
                  </div>

                  {/* Pending invitations */}
                  {pendingInvites.length > 0 && (
                    <div>
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
                        Pending Invitations
                      </p>
                      <div className="space-y-1">
                        {pendingInvites.map(inv => (
                          <div key={inv.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                            <div className="flex items-center space-x-2">
                              <ClockIcon className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                              <div>
                                <p className="text-sm text-foreground">{inv.email}</p>
                                <p className="text-xs text-muted-foreground">{roleLabel(inv.role)} · Invite pending</p>
                              </div>
                            </div>
                            <button
                              onClick={() => handleCancelInvitation(selectedTeam.id, inv.id)}
                              className="p-1.5 text-destructive hover:bg-destructive/10 rounded transition-colors"
                              title="Cancel invitation"
                            >
                              <XMarkIcon className="h-4 w-4" />
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Shared Links */}
                  <div className="mt-6 border-t border-border pt-4">
                    <div className="flex items-center space-x-2 mb-3">
                      <LinkIcon className="h-4 w-4 text-muted-foreground" />
                      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                        Shared Links
                      </p>
                    </div>
                    {linksLoading ? (
                      <div className="space-y-2">
                        {[1, 2].map(i => <div key={i} className="h-10 bg-muted rounded animate-pulse" />)}
                      </div>
                    ) : teamLinks.length === 0 ? (
                      <p className="text-sm text-muted-foreground">
                        No shared links yet. Create a link and assign it to this team.
                      </p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="text-xs text-muted-foreground border-b border-border">
                              <th className="text-left pb-2 font-medium">Short URL</th>
                              <th className="text-left pb-2 font-medium hidden sm:table-cell">Destination</th>
                              <th className="text-right pb-2 font-medium">Clicks</th>
                              <th className="text-right pb-2 font-medium hidden sm:table-cell">Created</th>
                            </tr>
                          </thead>
                          <tbody>
                            {teamLinks.map(link => (
                              <tr key={link.id} className="border-b border-border last:border-0">
                                <td className="py-2 pr-3">
                                  <a
                                    href={`/${link.short_code}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="text-primary hover:underline font-mono text-xs"
                                  >
                                    /{link.short_code}
                                  </a>
                                </td>
                                <td className="py-2 pr-3 hidden sm:table-cell max-w-[160px]">
                                  <p className="truncate text-xs text-muted-foreground" title={link.original_url}>
                                    {link.original_url}
                                  </p>
                                </td>
                                <td className="py-2 text-right text-xs font-medium">{link.click_count}</td>
                                <td className="py-2 text-right text-xs text-muted-foreground hidden sm:table-cell">
                                  {new Date(link.created_at).toLocaleDateString()}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Create Team Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-foreground">Create Team</h2>
              <button onClick={() => { setShowCreateModal(false); setError(''); }} className="p-1 text-muted-foreground hover:text-foreground rounded">
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
            <div className="space-y-3">
              <input
                type="text"
                value={teamName}
                onChange={e => setTeamName(e.target.value)}
                placeholder="Team name"
                className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
              <textarea
                value={teamDesc}
                onChange={e => setTeamDesc(e.target.value)}
                placeholder="Description (optional)"
                rows={2}
                className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
              />
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex gap-3 mt-4">
              <button onClick={() => { setShowCreateModal(false); setError(''); }} className="flex-1 px-4 py-2 text-sm text-foreground border border-border rounded-lg hover:bg-muted transition-colors">Cancel</button>
              <button onClick={handleCreateTeam} disabled={submitting} className="flex-1 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors">
                {submitting ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Invite Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-foreground">Invite Member</h2>
              <button onClick={() => { setShowInviteModal(false); setError(''); }} className="p-1 text-muted-foreground hover:text-foreground rounded">
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
            <div className="space-y-3">
              <input
                type="email"
                value={inviteEmail}
                onChange={e => setInviteEmail(e.target.value)}
                placeholder="colleague@example.com"
                className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
              <select
                value={inviteRole}
                onChange={e => setInviteRole(e.target.value)}
                className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              >
                <option value="member">Member</option>
                <option value="admin">Admin</option>
                <option value="viewer">Viewer</option>
              </select>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <div className="flex gap-3 mt-4">
              <button onClick={() => { setShowInviteModal(false); setError(''); }} className="flex-1 px-4 py-2 text-sm text-foreground border border-border rounded-lg hover:bg-muted transition-colors">Cancel</button>
              <button onClick={handleInvite} disabled={submitting} className="flex-1 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors">
                {submitting ? 'Sending...' : 'Send Invite'}
              </button>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
