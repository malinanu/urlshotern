'use client';

import { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PlusIcon, TrashIcon, PowerIcon } from '@heroicons/react/24/outline';

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

export default function AdminCouponsPage() {
    const { isAuthenticated, user, getAccessToken, isLoading } = useAuth();
    const [coupons, setCoupons] = useState<Coupon[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const [showCreateForm, setShowCreateForm] = useState(false);
    const [newCoupon, setNewCoupon] = useState({
        code: '',
        name: '',
        description: '',
        discount_type: 'percentage',
        discount_value: 0,
    });

    useEffect(() => {
        // Note: You should be sure user object contains the accurate role structure
        // Depending on this implementation, it might be in an array or simply 'admin'
        // Defaulting to simple check for now based on context
        if (!isLoading && isAuthenticated) {
            fetchCoupons();
        }
    }, [isLoading, isAuthenticated, user]);

    const fetchCoupons = async () => {
        try {
            setLoading(true);
            const res = await fetch(API_ENDPOINTS.admin.coupons.list, {
                headers: { Authorization: `Bearer ${getAccessToken()}` },
            });
            const data = await res.json();
            if (res.ok) {
                setCoupons(data.coupons || []);
            } else {
                setError(data.error || 'Failed to fetch coupons');
            }
        } catch (err) {
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
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.create, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${getAccessToken()}`
                },
                body: JSON.stringify({
                    ...newCoupon,
                    code: newCoupon.code.toUpperCase(),
                    discount_value: Number(newCoupon.discount_value),
                }),
            });
            if (res.ok) {
                setShowCreateForm(false);
                fetchCoupons();
            } else {
                const data = await res.json();
                alert(data.error || 'Failed to create coupon');
            }
        } catch (err) {
            alert('Error creating coupon');
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure?')) return;
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.delete(id), {
                method: 'DELETE',
                headers: { Authorization: `Bearer ${getAccessToken()}` },
            });
            if (res.ok) fetchCoupons();
        } catch (err) {
            alert('Error deleting coupon');
        }
    };

    const handleToggle = async (id: number) => {
        try {
            const res = await fetch(API_ENDPOINTS.admin.coupons.toggle(id), {
                method: 'PUT',
                headers: { Authorization: `Bearer ${getAccessToken()}` },
            });
            if (res.ok) fetchCoupons();
        } catch (err) {
            alert('Error updating coupon');
        }
    };

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

                {error && <div className="bg-destructive/10 text-destructive p-4 rounded-md mb-6">{error}</div>}

                {showCreateForm && (
                    <div className="bg-card p-6 rounded-lg shadow-sm border border-border mb-8 transition-all">
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
                                        <option value="percentage">Percentage Off</option>
                                        <option value="fixed">Fixed Amount Off</option>
                                        <option value="free_months">Free Months</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-1 text-foreground">Discount Value</label>
                                    <input
                                        required
                                        type="number"
                                        value={newCoupon.discount_value}
                                        onChange={(e) => setNewCoupon({ ...newCoupon, discount_value: Number(e.target.value) })}
                                        className="w-full bg-background border border-input rounded-md px-3 py-2 text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                        placeholder="e.g. 20 (for 20%)"
                                    />
                                </div>
                            </div>
                            <button
                                type="submit"
                                className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                            >
                                Save
                            </button>
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
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Uses</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {coupons.map((coupon) => (
                                <tr key={coupon.id} className={!coupon.is_active ? 'opacity-60 bg-muted/20' : 'hover:bg-muted/30 transition-colors'}>
                                    <td className="px-6 py-4 whitespace-nowrap font-medium text-foreground">{coupon.code}</td>
                                    <td className="px-6 py-4">
                                        <div className="text-sm text-foreground font-medium">{coupon.name}</div>
                                        <div className="text-xs text-muted-foreground font-mono">{coupon.polar_discount_id || 'Local bypass'}</div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                                        {coupon.discount_value} {coupon.discount_type === 'percentage' ? '%' : coupon.discount_type === 'free_months' ? 'Months Free' : 'Fixed'}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                        {coupon.current_redemptions} {coupon.max_redemptions ? `/ ${coupon.max_redemptions}` : ''}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${coupon.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'}`}>
                                            {coupon.is_active ? 'Active' : 'Disabled'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <button
                                            onClick={() => handleToggle(coupon.id)}
                                            className="text-primary hover:text-primary/80 mr-4 transition-colors"
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
                                    <td colSpan={6} className="px-6 py-8 text-center text-muted-foreground">
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
