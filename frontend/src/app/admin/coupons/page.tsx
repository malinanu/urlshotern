'use client';

import { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PlusIcon, TrashIcon, PowerIcon, ArrowPathIcon, CheckCircleIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';

interface Coupon {
    id: number;
    code: string;
    name: string;
    description: string;
    discount_type: string;
    discount_value: number;
    max_redemptions?: number;
    current_redemptions: number;
    valid_until?: string;
    is_active: boolean;
    polar_discount_id?: string;
}

function discountLabel(coupon: Coupon): string {
    if (coupon.discount_type === 'percentage') return `${coupon.discount_value}% off`;
    if (coupon.discount_type === 'fixed') return `$${coupon.discount_value} off`;
    return String(coupon.discount_value);
}

function valueLabel(type: string): string {
    if (type === 'percentage') return 'Percentage (1–100, e.g. 20 = 20% off)';
    if (type === 'fixed') return 'Amount in USD (e.g. 5 = $5 off)';
    return 'Value';
}

function valuePlaceholder(type: string): string {
    if (type === 'percentage') return 'e.g. 20';
    if (type === 'fixed') return 'e.g. 5';
    return 'e.g. 10';
}

export default function AdminCouponsPage() {
    const { isAuthenticated, user, getAccessToken, isLoading } = useAuth();
    const [coupons, setCoupons] = useState<Coupon[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [polarWarning, setPolarWarning] = useState('');
    const [syncing, setSyncing] = useState<number | null>(null);

    const [showCreateForm, setShowCreateForm] = useState(false);
    const [newCoupon, setNewCoupon] = useState({
        code: '',
        name: '',
        description: '',
        discount_type: 'percentage',
        discount_value: 0,
    });

    useEffect(() => {
        if (!isLoading && isAuthenticated) {
            fetchCoupons();
        }
    }, [isLoading, isAuthenticated, user]);

    const authHeader = () => ({ Authorization: `Bearer ${getAccessToken()}` });

    const fetchCoupons = async () => {
        try {
            setLoading(true);
            const res = await fetch(API_ENDPOINTS.admin.coupons.list, { headers: authHeader() });
            const data = await res.json();
            if (res.ok) {
                setCoupons(data.coupons || []);
            } else {
                setError(data.error || 'Failed to fetch coupons');
            }
        } catch {
            setError('Error fetching coupons');
        } finally {
            setLoading(false);
        }
    };

    if (isLoading || loading) {
        return <Layout><div className="p-8">Loading...</div></Layout>;
    }

    if (!isAuthenticated) {
        return <Layout><div className="p-8">Access Denied</div></Layout>;
    }

    const handleCreateCoupon = async (e: React.FormEvent) => {
        e.preventDefault();
        setPolarWarning('');
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.create, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', ...authHeader() },
                body: JSON.stringify({
                    ...newCoupon,
                    code: newCoupon.code.toUpperCase(),
                    discount_value: Number(newCoupon.discount_value),
                }),
            });
            const data = await res.json();
            if (res.ok) {
                setShowCreateForm(false);
                setNewCoupon({ code: '', name: '', description: '', discount_type: 'percentage', discount_value: 0 });
                if (data.polar_warning) setPolarWarning(data.polar_warning);
                fetchCoupons();
            } else {
                alert(data.error || 'Failed to create coupon');
            }
        } catch {
            alert('Error creating coupon');
        }
    };

    const handleSyncToPolar = async (coupon: Coupon) => {
        setSyncing(coupon.id);
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.syncPolar(coupon.id), {
                method: 'POST',
                headers: authHeader(),
            });
            const data = await res.json();
            if (res.ok) {
                setPolarWarning('');
                fetchCoupons();
            } else {
                alert(`Polar sync failed: ${data.details || data.error}`);
            }
        } catch {
            alert('Error syncing to Polar');
        } finally {
            setSyncing(null);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure?')) return;
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.delete(id), {
                method: 'DELETE',
                headers: authHeader(),
            });
            if (res.ok) fetchCoupons();
        } catch {
            alert('Error deleting coupon');
        }
    };

    const handleToggle = async (id: number) => {
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.toggle(id), {
                method: 'PUT',
                headers: authHeader(),
            });
            if (res.ok) fetchCoupons();
        } catch {
            alert('Error updating coupon');
        }
    };

    const needsPolarSync = (c: Coupon) =>
        (c.discount_type === 'percentage' || c.discount_type === 'fixed') && !c.polar_discount_id;

    return (
        <Layout>
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="flex justify-between items-center mb-6">
                    <h1 className="text-2xl font-bold text-foreground">Manage Discount Codes</h1>
                    <button
                        onClick={() => setShowCreateForm(!showCreateForm)}
                        className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                    >
                        <PlusIcon className="w-5 h-5 mr-2" />
                        {showCreateForm ? 'Cancel' : 'Add Coupon'}
                    </button>
                </div>

                {error && <div className="bg-destructive/10 text-destructive p-4 rounded-md mb-4">{error}</div>}

                {polarWarning && (
                    <div className="flex items-start gap-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-700 text-yellow-800 dark:text-yellow-300 p-4 rounded-md mb-4">
                        <ExclamationTriangleIcon className="h-5 w-5 mt-0.5 flex-shrink-0" />
                        <p className="text-sm">{polarWarning}</p>
                    </div>
                )}

                {showCreateForm && (
                    <div className="bg-card p-6 rounded-lg shadow-sm border border-border mb-8">
                        <h2 className="text-xl font-semibold mb-4 text-foreground">Create New Coupon</h2>
                        <form onSubmit={handleCreateCoupon} className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium mb-1 text-foreground">Code</label>
                                    <input
                                        required
                                        type="text"
                                        value={newCoupon.code}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, code: e.target.value.toUpperCase() })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary uppercase"
                                        placeholder="e.g. SUMMER20"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-1 text-foreground">Name</label>
                                    <input
                                        required
                                        type="text"
                                        value={newCoupon.name}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, name: e.target.value })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                        placeholder="Summer Sale 20%"
                                    />
                                </div>
                                <div className="md:col-span-2">
                                    <label className="block text-sm font-medium mb-1 text-foreground">Description</label>
                                    <input
                                        type="text"
                                        value={newCoupon.description}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, description: e.target.value })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                        placeholder="20% off all plans"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-1 text-foreground">Discount Type</label>
                                    <select
                                        value={newCoupon.discount_type}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, discount_type: e.target.value })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                    >
                                        <option value="percentage">Percentage Off — synced to Polar</option>
                                        <option value="fixed">Fixed Amount Off (USD) — synced to Polar</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-1 text-foreground">
                                        Value <span className="text-muted-foreground font-normal">({valueLabel(newCoupon.discount_type)})</span>
                                    </label>
                                    <input
                                        required
                                        type="number"
                                        min="1"
                                        value={newCoupon.discount_value || ''}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, discount_value: Number(e.target.value) })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                        placeholder={valuePlaceholder(newCoupon.discount_type)}
                                    />
                                </div>
                            </div>
                            <div className="flex items-center gap-3">
                                <button
                                    type="submit"
                                    className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                                >
                                    Save
                                </button>
                                {(newCoupon.discount_type === 'percentage' || newCoupon.discount_type === 'fixed') && (
                                    <p className="text-xs text-muted-foreground">
                                        This coupon will be automatically created in your Polar dashboard.
                                    </p>
                                )}
                            </div>
                        </form>
                    </div>
                )}

                <div className="bg-card rounded-lg border border-border overflow-hidden shadow-sm">
                    <table className="min-w-full divide-y divide-border">
                        <thead className="bg-muted/50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Code</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Name</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Discount</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Polar</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Uses</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {coupons.map((coupon) => (
                                <tr key={coupon.id} className={!coupon.is_active ? 'opacity-60 bg-muted/20' : 'hover:bg-muted/30 transition-colors'}>
                                    <td className="px-6 py-4 whitespace-nowrap font-mono font-medium text-foreground">{coupon.code}</td>
                                    <td className="px-6 py-4">
                                        <div className="text-sm text-foreground font-medium">{coupon.name}</div>
                                        {coupon.description && <div className="text-xs text-muted-foreground">{coupon.description}</div>}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                                        {discountLabel(coupon)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        {coupon.polar_discount_id ? (
                                            <span className="inline-flex items-center gap-1 text-xs text-green-700 dark:text-green-400">
                                                <CheckCircleIcon className="h-3.5 w-3.5" />
                                                Synced
                                            </span>
                                        ) : (
                                            <button
                                                onClick={() => handleSyncToPolar(coupon)}
                                                disabled={syncing === coupon.id}
                                                title="Click to sync this coupon to your Polar dashboard"
                                                className="inline-flex items-center gap-1 text-xs text-yellow-700 dark:text-yellow-400 hover:text-yellow-900 dark:hover:text-yellow-200 transition-colors disabled:opacity-50"
                                            >
                                                <ArrowPathIcon className={`h-3.5 w-3.5 ${syncing === coupon.id ? 'animate-spin' : ''}`} />
                                                {syncing === coupon.id ? 'Syncing…' : 'Sync to Polar'}
                                            </button>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                        {coupon.current_redemptions}{coupon.max_redemptions ? ` / ${coupon.max_redemptions}` : ''}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${coupon.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'}`}>
                                            {coupon.is_active ? 'Active' : 'Disabled'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-3">
                                        <button
                                            onClick={() => handleToggle(coupon.id)}
                                            className="text-primary hover:text-primary/80 transition-colors"
                                            title="Toggle Active Status"
                                        >
                                            <PowerIcon className="w-5 h-5 inline" />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(coupon.id)}
                                            className="text-destructive hover:text-destructive/80 transition-colors"
                                            title="Delete Coupon"
                                        >
                                            <TrashIcon className="w-5 h-5 inline" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                            {coupons.length === 0 && (
                                <tr>
                                    <td colSpan={7} className="px-6 py-8 text-center text-muted-foreground">
                                        No discount codes found.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </Layout>
    );
}
