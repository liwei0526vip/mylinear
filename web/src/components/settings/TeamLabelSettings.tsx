/**
 * 团队标签设置组件 - Premium 版本
 */

import { useState, useEffect } from 'react';
import { useLabelStore } from '../../stores/labelStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Plus, Trash2, Tag, Loader2, Info, Palette, MoreHorizontal, Check, X } from 'lucide-react';
import { cn } from '../../lib/utils';
import type { Label as LabelModel } from '../../types/label';

interface TeamLabelSettingsProps {
    teamId: string;
}

export function TeamLabelSettings({ teamId }: TeamLabelSettingsProps) {
    const { labels, loading, error, fetchLabels, addLabel, updateLabel, deleteLabel } = useLabelStore();

    // 新建标签状态
    const [isAdding, setIsAdding] = useState(false);
    const [newLabel, setNewLabel] = useState({ name: '', color: '#3b82f6' });

    // 编辑标签状态
    const [editingId, setEditingId] = useState<string | null>(null);
    const [editForm, setEditForm] = useState<{ name: string; color: string } | null>(null);

    useEffect(() => {
        if (teamId) {
            fetchLabels(teamId);
        }
    }, [teamId, fetchLabels]);

    const handleCreate = async () => {
        if (!newLabel.name) return;
        try {
            await addLabel(teamId, newLabel);
            setNewLabel({ name: '', color: '#3b82f6' });
            setIsAdding(false);
        } catch (err) {
            console.error(err);
        }
    };

    const startEditing = (label: LabelModel) => {
        setEditingId(label.id);
        setEditForm({ name: label.name, color: label.color });
    };

    const handleUpdate = async (id: string) => {
        if (!editForm || !editForm.name) return;
        try {
            await updateLabel(id, editForm);
            setEditingId(null);
        } catch (err) {
            console.error(err);
        }
    };

    // 区分并排序 (首字母排序)
    const teamLabels = labels.filter(l => l.team_id === teamId).sort((a, b) => a.name.localeCompare(b.name));
    const workspaceLabels = labels.filter(l => !l.team_id).sort((a, b) => a.name.localeCompare(b.name));

    return (
        <div className="space-y-8 animate-in fade-in duration-500">

            {/* 1. 团队标签卡片 */}
            <Card className="overflow-hidden border-primary/5 shadow-sm">
                <CardHeader className="flex flex-row items-center justify-between bg-accent/20 border-b border-primary/5 py-5 px-6">
                    <div className="space-y-1">
                        <div className="flex items-center gap-2">
                            <Tag className="h-4 w-4 text-primary mt-0.5" />
                            <CardTitle className="text-lg">团队专属标签</CardTitle>
                        </div>
                        <CardDescription className="text-xs">
                            创建仅在 {teamId.substring(0, 5)}... 内部可见的标签
                        </CardDescription>
                    </div>
                    <Button
                        size="sm"
                        onClick={() => setIsAdding(true)}
                        disabled={isAdding}
                        className="shadow-sm"
                    >
                        <Plus className="mr-2 h-4 w-4" />
                        添加标签
                    </Button>
                </CardHeader>
                <CardContent className="p-6">
                    {error && (
                        <div className="mb-6 rounded-lg bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2 border border-destructive/20">
                            <Info className="h-4 w-4" />
                            {error}
                        </div>
                    )}

                    {loading && labels.length === 0 ? (
                        <div className="py-10 flex flex-col items-center justify-center text-muted-foreground">
                            <Loader2 className="h-8 w-8 animate-spin mb-2 opacity-30" />
                            <p className="text-xs font-medium uppercase tracking-widest">同步中...</p>
                        </div>
                    ) : (
                        <div className="flex flex-wrap gap-2.5">
                            {teamLabels.map((label) => (
                                <div key={label.id} className="relative group">
                                    {editingId === label.id ? (
                                        <div className="flex items-center gap-2 rounded-full border-2 border-primary/40 bg-background px-2 py-0.5 animate-in zoom-in-95">
                                            <input
                                                type="color"
                                                className="w-4 h-4 rounded-full border-none p-0 bg-transparent cursor-pointer"
                                                value={editForm?.color}
                                                onChange={e => setEditForm(prev => prev ? { ...prev, color: e.target.value } : null)}
                                            />
                                            <Input
                                                className="h-7 w-24 border-none bg-transparent p-0 text-xs focus-visible:ring-0 font-medium"
                                                value={editForm?.name}
                                                onChange={e => setEditForm(prev => prev ? { ...prev, name: e.target.value } : null)}
                                                autoFocus
                                                onKeyDown={e => e.key === 'Enter' && handleUpdate(label.id)}
                                            />
                                            <button className="text-primary hover:bg-primary/10 p-1 rounded" onClick={() => handleUpdate(label.id)}>
                                                <Check className="h-3 w-3" />
                                            </button>
                                            <button className="text-muted-foreground hover:bg-destructive/10 p-1 rounded" onClick={() => setEditingId(null)}>
                                                <X className="h-3 w-3" />
                                            </button>
                                        </div>
                                    ) : (
                                        <div
                                            className={cn(
                                                "flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-medium cursor-pointer transition-all hover:shadow-md",
                                                "bg-background hover:border-primary/40"
                                            )}
                                            onClick={() => startEditing(label)}
                                        >
                                            <div className="h-2 w-2 rounded-full shadow-sm" style={{ backgroundColor: label.color }} />
                                            <span>{label.name}</span>
                                            <button
                                                className="ml-1 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-all p-0.5 rounded-full hover:bg-destructive/5"
                                                onClick={(e) => { e.stopPropagation(); deleteLabel(label.id); }}
                                            >
                                                <Trash2 className="h-3 w-3" />
                                            </button>
                                        </div>
                                    )}
                                </div>
                            ))}

                            {isAdding && (
                                <div className="flex items-center gap-2 rounded-full border-2 border-dashed border-primary/50 bg-primary/5 px-2 py-0.5 animate-in slide-in-from-left-2 shadow-sm">
                                    <div className="relative group">
                                        <input
                                            type="color"
                                            className="h-6 w-6 rounded-full border-2 border-background bg-transparent p-0 cursor-pointer overflow-hidden"
                                            value={newLabel.color}
                                            onChange={(e) => setNewLabel({ ...newLabel, color: e.target.value })}
                                        />
                                    </div>
                                    <Input
                                        className="h-7 w-28 border-none bg-transparent p-0 text-xs focus-visible:ring-0 font-semibold placeholder:font-normal"
                                        value={newLabel.name}
                                        onChange={(e) => setNewLabel({ ...newLabel, name: e.target.value })}
                                        placeholder="新标签名"
                                        autoFocus
                                        onKeyDown={(e) => {
                                            if (e.key === 'Enter') handleCreate();
                                            if (e.key === 'Escape') setIsAdding(false);
                                        }}
                                    />
                                    <div className="flex gap-1 pr-1">
                                        <button className="p-1 rounded bg-primary text-primary-foreground" onClick={handleCreate}>
                                            <Check className="h-3 w-3" />
                                        </button>
                                        <button className="p-1 rounded bg-muted text-muted-foreground" onClick={() => setIsAdding(false)}>
                                            <X className="h-3 w-3" />
                                        </button>
                                    </div>
                                </div>
                            )}

                            {teamLabels.length === 0 && !isAdding && (
                                <div className="w-full py-8 text-center border-2 border-dashed rounded-xl bg-muted/5">
                                    <p className="text-sm text-muted-foreground font-medium">暂无专属标签</p>
                                    <p className="text-[10px] text-muted-foreground mt-1 px-10">点击上方按钮创建第一个仅供本团队使用的任务标签</p>
                                </div>
                            )}
                        </div>
                    )}
                </CardContent>
            </Card>

            {/* 2. 工作区标签卡片 (Preview Only) */}
            <Card className="border-primary/5 bg-accent/5 overflow-hidden">
                <CardHeader className="py-4 px-6 border-b border-primary/5 bg-background/50">
                    <div className="flex items-center gap-2">
                        <Palette className="h-4 w-4 text-muted-foreground" />
                        <CardTitle className="text-base text-muted-foreground">全局共享标签</CardTitle>
                    </div>
                    <CardDescription className="text-[10px]">
                        工作区预定义的全局标签，所有团队通用，此处仅供参考，不可修改
                    </CardDescription>
                </CardHeader>
                <CardContent className="p-6">
                    <div className="flex flex-wrap gap-2.5 opacity-60 grayscale-[0.2]">
                        {workspaceLabels.map((label) => (
                            <div
                                key={label.id}
                                className="flex items-center gap-2 rounded-md border border-muted bg-background/50 px-2.5 py-1.5 text-xs font-medium cursor-not-allowed"
                                title="这是全局标签，无法在团队设置中修改"
                            >
                                <div className="h-2 w-2 rounded-full" style={{ backgroundColor: label.color }} />
                                <span>{label.name}</span>
                                <Info className="h-2.5 w-2.5 ml-1 opacity-20" />
                            </div>
                        ))}

                        {workspaceLabels.length === 0 && (
                            <p className="text-xs text-muted-foreground italic">工作区暂未定义任何全局标签</p>
                        )}
                    </div>
                </CardContent>
            </Card>

        </div>
    );
}
