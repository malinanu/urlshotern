'use client';

import { useEffect, useState, useRef, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import {
  MegaphoneIcon,
  PlusIcon,
  UsersIcon,
  PaperAirplaneIcon,
  ChartBarIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface Campaign {
  id: number;
  name: string;
  description: string;
  ai_context: string;
  channels: string[];
  status: string;
  created_at: string;
}

interface CampaignStats {
  campaign_id: number;
  total: number;
  sent: number;
  failed: number;
  joined: number;
}

export default function CampaignsPage() {
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const router = useRouter();
  
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [statsMap, setStatsMap] = useState<Record<number, CampaignStats>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal states
  const [isNewModalOpen, setIsNewModalOpen] = useState(false);
  const [isContactsModalOpen, setIsContactsModalOpen] = useState(false);
  const [selectedCampaignId, setSelectedCampaignId] = useState<number | null>(null);

  // New Campaign Form
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newContext, setNewContext] = useState('');
  const [newChannels, setNewChannels] = useState<string[]>(['email']);

  // CSV Import
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchCampaigns = useCallback(async () => {
    if (!isAuthenticated) return;
    const token = getAccessToken();
    if (!token) return;

    try {
      setLoading(true);
      const res = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns?limit=50`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Failed to fetch campaigns');
      const data = await res.json();
      setCampaigns(data.campaigns || []);

      // Fetch stats for all campaigns
      const newStatsMap: Record<number, CampaignStats> = {};
      for (const camp of (data.campaigns || [])) {
        try {
          const statsRes = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns/${camp.id}/stats`, {
            headers: { Authorization: `Bearer ${token}` }
          });
          if (statsRes.ok) {
            newStatsMap[camp.id] = await statsRes.json();
          }
        } catch (e) {
          console.error('Failed to fetch stats for campaign', camp.id);
        }
      }
      setStatsMap(newStatsMap);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated, getAccessToken]);

  useEffect(() => {
    fetchCampaigns();
  }, [fetchCampaigns]);

  const handleCreateCampaign = async (e: React.FormEvent) => {
    e.preventDefault();
    const token = getAccessToken();
    if (!token) return;

    try {
      const payload = {
        name: newName,
        description: newDesc,
        ai_context: newContext,
        channels: newChannels
      };
      const res = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });
      if (!res.ok) throw new Error('Failed to create campaign');
      
      setIsNewModalOpen(false);
      setNewName(''); setNewDesc(''); setNewContext(''); setNewChannels(['email']);
      fetchCampaigns();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleToggleChannel = (ch: string) => {
    setNewChannels(prev => 
      prev.includes(ch) ? prev.filter(c => c !== ch) : [...prev, ch]
    );
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this campaign and all its data?')) return;
    const token = getAccessToken();
    if (!token) return;

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) throw new Error('Failed to delete campaign');
      fetchCampaigns();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleSend = async (id: number) => {
    if (!confirm('This will trigger the campaign and send messages. Continue?')) return;
    const token = getAccessToken();
    if (!token) return;

    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns/${id}/send`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }
      });
      if (!res.ok) {
        const d = await res.json().catch(()=>({}));
        throw new Error(d.error || 'Failed to send campaign');
      }
      alert('Campaign send started successfully.');
      fetchCampaigns();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !selectedCampaignId) return;
    const token = getAccessToken();
    if (!token) return;

    const formData = new FormData();
    formData.append('file', file);

    setUploading(true);
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/admin/campaigns/${selectedCampaignId}/contacts`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` }, // Note: Don't set Content-Type for FormData
        body: formData
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to upload contacts');
      
      alert(`Successfully imported ${data.imported} contacts!`);
      setIsContactsModalOpen(false);
      fetchCampaigns();
    } catch (err: any) {
      alert(err.message);
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  // Ensure only admins can see this
  if (!user?.is_admin) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center bg-background text-destructive text-xl font-bold">
          403 Forbidden - Admin Only
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center mb-8">
            <div>
              <h1 className="text-3xl font-bold text-foreground">Promotional Campaigns</h1>
              <p className="mt-2 text-muted-foreground">Manage and track AI-driven outbound marketing.</p>
            </div>
            <button
              onClick={() => setIsNewModalOpen(true)}
              className="flex items-center space-x-2 bg-primary text-primary-foreground px-4 py-2 rounded-md hover:bg-primary/90 transition-colors"
            >
              <PlusIcon className="h-5 w-5" />
              <span>New Campaign</span>
            </button>
          </div>

          {loading ? (
            <div className="animate-pulse space-y-4">
              {[1,2,3].map(i => (
                <div key={i} className="h-24 bg-muted rounded-lg w-full"></div>
              ))}
            </div>
          ) : error ? (
            <div className="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-md">
              {error}
            </div>
          ) : campaigns.length === 0 ? (
            <div className="text-center py-16 bg-card border border-border rounded-lg shadow-sm">
              <MegaphoneIcon className="mx-auto h-16 w-16 text-muted-foreground opacity-50" />
              <h3 className="mt-4 text-lg font-medium text-foreground">No campaigns yet</h3>
              <p className="mt-2 text-muted-foreground max-w-md mx-auto">
                Create a campaign to send AI-generated promotions via Email, SMS, or WhatsApp and track conversions.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6">
              {campaigns.map(camp => {
                const stats = statsMap[camp.id];
                const convRate = stats?.total > 0 ? Math.round((stats.joined / stats.total) * 100) : 0;
                return (
                  <div key={camp.id} className="bg-card border border-border rounded-lg shadow-sm p-6 flex flex-col md:flex-row md:items-center justify-between gap-4 group">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <h2 className="text-xl font-bold text-foreground">{camp.name}</h2>
                        <span className={`px-2.5 py-0.5 rounded-full text-xs font-semibold uppercase tracking-wide
                          ${camp.status === 'draft' ? 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-300' :
                            camp.status === 'sending' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300' :
                            'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'}`}>
                          {camp.status}
                        </span>
                      </div>
                      <p className="text-muted-foreground text-sm mb-3">{camp.description}</p>
                      <div className="flex items-center space-x-2 text-sm text-foreground mb-4 md:mb-0">
                        <span className="font-semibold text-muted-foreground text-xs uppercase tracking-wider mr-2">Channels:</span>
                        {camp.channels.map(ch => (
                          <span key={ch} className="bg-muted px-2 py-0.5 rounded text-xs">{ch}</span>
                        ))}
                      </div>
                    </div>

                    <div className="flex-1 flex flex-col md:flex-row md:items-center md:justify-end gap-6 shrink-0">
                      {/* Stats */}
                      {stats && (
                        <div className="flex space-x-6 text-sm">
                          <div className="text-center">
                            <p className="text-muted-foreground text-xs uppercase">Contacts</p>
                            <p className="font-bold text-xl text-foreground">{stats.total}</p>
                          </div>
                          <div className="text-center">
                            <p className="text-muted-foreground text-xs uppercase">Sent</p>
                            <p className="font-bold text-xl text-blue-500">{stats.sent}</p>
                          </div>
                          <div className="text-center">
                            <p className="text-muted-foreground text-xs uppercase">Joined</p>
                            <div className="flex items-end justify-center space-x-1">
                              <p className="font-bold text-xl text-green-500">{stats.joined}</p>
                              <span className="text-xs text-green-600 mb-1">({convRate}%)</span>
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Actions */}
                      <div className="flex flex-wrap items-center gap-2 border-t md:border-t-0 md:border-l border-border pt-4 md:pt-0 md:pl-6">
                        <button
                          onClick={() => {
                            setSelectedCampaignId(camp.id);
                            setIsContactsModalOpen(true);
                          }}
                          className="flex items-center space-x-1 px-3 py-1.5 bg-muted hover:bg-muted/80 text-foreground text-sm rounded-md transition-colors"
                          disabled={camp.status !== 'draft'}
                        >
                          <UsersIcon className="w-4 h-4" />
                          <span>Import</span>
                        </button>
                        <button
                          onClick={() => handleSend(camp.id)}
                          className="flex items-center space-x-1 px-3 py-1.5 bg-primary text-primary-foreground hover:bg-primary/90 text-sm rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          disabled={camp.status !== 'draft' || !stats || stats.total === 0}
                        >
                          <PaperAirplaneIcon className="w-4 h-4" />
                          <span>Send</span>
                        </button>
                        <button
                          onClick={() => handleDelete(camp.id)}
                          className="p-1.5 text-muted-foreground hover:text-destructive transition-colors rounded-md hover:bg-muted"
                          title="Delete Campaign"
                        >
                          <TrashIcon className="w-5 h-5" />
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* New Campaign Modal */}
      {isNewModalOpen && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-card w-full max-w-lg border border-border rounded-xl shadow-xl overflow-hidden flex flex-col max-h-[90vh]">
            <div className="p-6 border-b border-border">
              <h2 className="text-xl font-bold text-foreground">Create Campaign</h2>
            </div>
            <form onSubmit={handleCreateCampaign} className="flex flex-col overflow-y-auto p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-1">Campaign Name</label>
                <input required value={newName} onChange={e=>setNewName(e.target.value)} className="w-full p-2 border border-input rounded bg-background" placeholder="e.g. Summer Promo" />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1">Description</label>
                <input required value={newDesc} onChange={e=>setNewDesc(e.target.value)} className="w-full p-2 border border-input rounded bg-background" placeholder="Internal notes" />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1">AI Context Prompt</label>
                <p className="text-xs text-muted-foreground mb-2">Instructions for the AI that will generate the email body. Explain the goal, the tone, and what offering to pitch.</p>
                <textarea required rows={4} value={newContext} onChange={e=>setNewContext(e.target.value)} className="w-full p-2 border border-input rounded bg-background" placeholder="e.g. We are a link shortener offering a 50% discount on the Pro plan..." />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">Channels</label>
                <div className="flex space-x-4">
                  {['email', 'sms', 'whatsapp'].map(ch => (
                    <label key={ch} className="flex items-center space-x-2 cursor-pointer">
                      <input 
                        type="checkbox" 
                        checked={newChannels.includes(ch)} 
                        onChange={() => handleToggleChannel(ch)}
                        className="rounded border-input text-primary focus:ring-primary"
                      />
                      <span className="capitalize">{ch}</span>
                    </label>
                  ))}
                </div>
              </div>
              <div className="mt-6 flex justify-end space-x-3 pt-4 border-t border-border">
                <button type="button" onClick={()=>setIsNewModalOpen(false)} className="px-4 py-2 border border-input rounded hover:bg-muted text-sm font-medium">Cancel</button>
                <button type="submit" disabled={newChannels.length === 0} className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 text-sm font-medium disabled:opacity-50">Create Campaign</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Import Contacts Modal */}
      {isContactsModalOpen && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-card w-full max-w-sm border border-border rounded-xl shadow-xl overflow-hidden p-6 text-center">
            <h2 className="text-xl font-bold mb-4">Import Contacts</h2>
            <p className="text-sm text-muted-foreground mb-6">Upload a CSV file containing <code className="bg-muted px-1 rounded text-red-500 font-mono">name</code>, <code className="bg-muted px-1 rounded text-red-500 font-mono">email</code>, and <code className="bg-muted px-1 rounded text-red-500 font-mono">phone</code> columns.</p>
            
            <input 
              type="file" 
              accept=".csv"
              ref={fileInputRef}
              onChange={handleFileUpload}
              className="hidden" 
              id="csv-upload"
            />
            <label 
              htmlFor="csv-upload" 
              className={`cursor-pointer block w-full py-6 border-2 border-dashed border-input rounded-lg hover:border-primary hover:bg-primary/5 transition-colors ${uploading ? 'opacity-50 pointer-events-none' : ''}`}
            >
              {uploading ? (
                <div className="flex flex-col items-center">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mb-2"></div>
                  <span>Uploading...</span>
                </div>
              ) : (
                <div className="flex flex-col items-center">
                  <PaperAirplaneIcon className="h-8 w-8 text-muted-foreground mb-2 rotate-90" />
                  <span className="text-sm font-medium">Click to select CSV</span>
                </div>
              )}
            </label>

            <button onClick={() => setIsContactsModalOpen(false)} disabled={uploading} className="mt-6 text-sm text-muted-foreground hover:text-foreground">
              Cancel
            </button>
          </div>
        </div>
      )}

    </Layout>
  );
}
