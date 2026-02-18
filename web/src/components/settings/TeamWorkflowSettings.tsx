/**
 * 团队工作流设置组件 - 支持编辑功能
 */

import { useState, useEffect } from 'react';
import { useWorkflowStore } from '../../stores/workflowStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Plus, Trash2, Loader2, Info, Palette, Edit2 } from 'lucide-react';
import type { StateType, WorkflowState } from '../../types/workflow';

interface TeamWorkflowSettingsProps {
    teamId: string;
}

export function TeamWorkflowSettings({ teamId }: TeamWorkflowSettingsProps) {
    const { states, loading, error, fetchStates, addState, updateState, deleteState } = useWorkflowStore();

    // 添加状态相关
    const [isAdding, setIsAdding] = useState(false);
    const [newState, setNewState] = useState({
        name: '',
        type: 'unstarted' as StateType,
        color: '#cbd5e1',
    });

    // 编辑状态相关
    const [editingId, setEditingId] = useState<string | null>(null);
    const [editForm, setEditForm] = useState<{
        name: string;
        type: StateType;
        color: string;
    } | null>(null);

    useEffect(() => {
        if (teamId) {
            fetchStates(teamId);
        }
    }, [teamId, fetchStates]);

    const handleAdd = async () => {
        if (!newState.name) return;
        try {
            await addState(teamId, {
                ...newState,
                position: states.length > 0 ? Math.max(...states.map(s => s.position)) + 10 : 10,
            });
            setNewState({ name: '', type: 'unstarted', color: '#cbd5e1' });
            setIsAdding(false);
        } catch (err) {
            console.error(err);
        }
    };

    const startEditing = (state: WorkflowState) => {
        setEditingId(state.id);
        setEditForm({
            name: state.name,
            type: state.type,
            color: state.color,
        });
    };

    const handleUpdate = async (id: string) => {
        if (!editForm || !editForm.name) return;
        try {
            await updateState(id, editForm);
            setEditingId(null);
            setEditForm(null);
        } catch (err) {
            console.error(err);
        }
    };

    const getTypeLabel = (type: StateType) => {
        switch (type) {
            case 'backlog': return '待办';
            case 'unstarted': return '未开始';
            case 'started': return '进行中';
            case 'completed': return '已完成';
            case 'canceled': return '已取消';
            default: return type;
        }
    };

    const presetColors = ['#cbd5e1', '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#64748b'];

    return (
        <div className="space-y-6 animate-in fade-in duration-500">
            <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                    <div>
                        <CardTitle>工作流状态</CardTitle>
                        <CardDescription>管理团队任务的状态流转，类别决定了进度的自动统计</CardDescription>
                    </div>
                    {!isAdding && states.length > 0 && (
                        <Button size="sm" onClick={() => setIsAdding(true)}>
                            <Plus className="mr-2 h-4 w-4" />
                            添加状态
                        </Button>
                    )}
                </CardHeader>
                <CardContent>
                    {error && (
                        <div className="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                            <Info className="h-4 w-4" />
                            {error}
                        </div>
                    )}

                    {loading && states.length === 0 ? (
                        <div className="py-12 flex flex-col items-center justify-center text-muted-foreground">
                            <Loader2 className="h-8 w-8 animate-spin mb-2" />
                            <p className="text-sm">正在获取工作流配置...</p>
                        </div>
                    ) : (
                        <div className="space-y-3">
                            {/* 状态列表 */}
                            <div className="space-y-3">
                                {states.map((state) => (
                                    <div key={state.id} className="group relative">
                                        {editingId === state.id ? (
                                            // 编辑模式表单
                                            <div className="rounded-lg border-2 border-primary/30 p-4 bg-accent/10 space-y-4 animate-in zoom-in-95 duration-200 shadow-sm">
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                    <div className="space-y-1">
                                                        <Label className="text-[10px] font-bold uppercase tracking-wider opacity-60">状态名称</Label>
                                                        <Input
                                                            size={1}
                                                            className="h-9 text-sm"
                                                            value={editForm?.name}
                                                            onChange={e => setEditForm(prev => prev ? { ...prev, name: e.target.value } : null)}
                                                        />
                                                    </div>
                                                    <div className="space-y-1">
                                                        <Label className="text-[10px] font-bold uppercase tracking-wider opacity-60">状态类别</Label>
                                                        <select
                                                            className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background focus-visible:outline-none focus"
                                                            value={editForm?.type}
                                                            onChange={e => setEditForm(prev => prev ? { ...prev, type: e.target.value as StateType } : null)}
                                                        >
                                                            <option value="backlog">待办 (Backlog)</option>
                                                            <option value="unstarted">未开始 (Unstarted)</option>
                                                            <option value="started">进行中 (Started)</option>
                                                            <option value="completed">已完成 (Completed)</option>
                                                            <option value="canceled">已取消 (Canceled)</option>
                                                        </select>
                                                    </div>
                                                </div>
                                                <div className="flex items-center justify-between pt-2 border-t border-primary/5">
                                                    <div className="flex items-center gap-2">
                                                        <input
                                                            type="color"
                                                            className="h-7 w-7 rounded cursor-pointer border-none p-0 bg-transparent"
                                                            value={editForm?.color}
                                                            onChange={e => setEditForm(prev => prev ? { ...prev, color: e.target.value } : null)}
                                                        />
                                                        <div className="flex gap-1">
                                                            {presetColors.map(c => (
                                                                <button
                                                                    key={c}
                                                                    className={`w-4 h-4 rounded-full border ${editForm?.color === c ? 'ring-2 ring-primary ring-offset-1' : 'opacity-60'}`}
                                                                    style={{ backgroundColor: c }}
                                                                    onClick={() => setEditForm(prev => prev ? { ...prev, color: c } : null)}
                                                                />
                                                            ))}
                                                        </div>
                                                    </div>
                                                    <div className="flex gap-2">
                                                        <Button variant="ghost" size="sm" className="h-8 px-3" onClick={() => setEditingId(null)}>
                                                            取消
                                                        </Button>
                                                        <Button size="sm" className="h-8 px-3" onClick={() => handleUpdate(state.id)}>
                                                            保存修改
                                                        </Button>
                                                    </div>
                                                </div>
                                            </div>
                                        ) : (
                                            // 普通展示模式
                                            <div
                                                className="flex items-center justify-between rounded-lg border p-4 bg-card hover:bg-accent/40 hover:border-primary/20 transition-all cursor-default"
                                                onClick={() => startEditing(state)}
                                            >
                                                <div className="flex items-center gap-4">
                                                    <div
                                                        className="h-4 w-4 rounded-full shadow-sm"
                                                        style={{ backgroundColor: state.color }}
                                                    />
                                                    <div>
                                                        <div className="font-semibold text-sm">{state.name}</div>
                                                        <div className="text-[10px] text-muted-foreground uppercase tracking-widest font-medium">
                                                            {getTypeLabel(state.type)}
                                                        </div>
                                                    </div>
                                                </div>
                                                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 text-muted-foreground hover:text-primary"
                                                        onClick={(e) => { e.stopPropagation(); startEditing(state); }}
                                                    >
                                                        <Edit2 className="h-3.5 w-3.5" />
                                                    </Button>
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                                                        onClick={(e) => { e.stopPropagation(); deleteState(state.id); }}
                                                        disabled={states.filter(s => s.type === state.type).length <= 1}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>

                            {/* 空状态展示 */}
                            {states.length === 0 && !isAdding && (
                                <div className="py-16 flex flex-col items-center justify-center border-2 border-dashed rounded-xl bg-muted/10">
                                    <div className="bg-background p-4 rounded-full shadow-md mb-4 animate-bounce">
                                        <Plus className="h-6 w-6 text-primary" />
                                    </div>
                                    <h3 className="font-bold text-base">定制你的工作流</h3>
                                    <p className="text-sm text-muted-foreground mb-6 max-w-[280px] text-center">系统未能自动生成状态，请手动创建第一个状态来开始管理团队任务</p>
                                    <Button size="lg" onClick={() => setIsAdding(true)} className="px-8 shadow-lg shadow-primary/20">
                                        立即创建
                                    </Button>
                                </div>
                            )}

                            {/* 添加表单 */}
                            {isAdding && (
                                <div className="rounded-xl border-2 border-primary/20 p-6 space-y-6 bg-accent/10 animate-in fade-in slide-in-from-top-4 duration-300">
                                    <div className="flex items-center gap-3">
                                        <div className="p-2 bg-primary/10 rounded-lg">
                                            <Plus className="h-5 w-5 text-primary" />
                                        </div>
                                        <div>
                                            <h4 className="text-base font-bold text-primary">添加新状态</h4>
                                            <p className="text-xs text-muted-foreground">为你的团队增加一个新的流程节点</p>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <div className="space-y-3">
                                            <Label className="text-xs font-bold uppercase tracking-wider opacity-70 px-1">状态名称</Label>
                                            <Input
                                                className="h-10 text-sm bg-background border-primary/10 focus:border-primary/40 focus:ring-primary/10"
                                                value={newState.name}
                                                onChange={(e) => setNewState({ ...newState, name: e.target.value })}
                                                placeholder="例如：待审核、QA 测试中..."
                                                autoFocus
                                            />
                                        </div>
                                        <div className="space-y-3">
                                            <Label className="text-xs font-bold uppercase tracking-wider opacity-70 px-1">任务类别</Label>
                                            <select
                                                className="flex h-10 w-full rounded-md border border-primary/10 bg-background px-3 py-1 text-sm focus:border-primary/40 focus:ring-primary/10"
                                                value={newState.type}
                                                onChange={(e) => setNewState({ ...newState, type: e.target.value as StateType })}
                                            >
                                                <option value="backlog">待办 (Backlog) - 用于记录点子或长远规划</option>
                                                <option value="unstarted">未开始 (Unstarted) - 准备好要做的任务</option>
                                                <option value="started">进行中 (Started) - 实际正在开发中</option>
                                                <option value="completed">已完成 (Completed) - 开发并上线完成</option>
                                                <option value="canceled">已取消 (Canceled) - 确定不再执行的任务</option>
                                            </select>
                                        </div>
                                    </div>

                                    <div className="space-y-3 pt-2">
                                        <Label className="text-xs font-bold uppercase tracking-wider opacity-70 px-1 flex items-center gap-2">
                                            <Palette className="h-3.5 w-3.5" />
                                            标识颜色
                                        </Label>
                                        <div className="flex items-center gap-4 p-3 bg-background rounded-lg border border-primary/5">
                                            <input
                                                type="color"
                                                className="h-10 w-12 rounded cursor-pointer border-none p-0 bg-transparent"
                                                value={newState.color}
                                                onChange={(e) => setNewState({ ...newState, color: e.target.value })}
                                            />
                                            <span className="text-sm font-mono text-muted-foreground">{newState.color}</span>
                                            <div className="flex gap-2 ml-4 border-l pl-4">
                                                {presetColors.map(c => (
                                                    <button
                                                        key={c}
                                                        className={`w-6 h-6 rounded-full border-2 border-background shadow-sm transition-transform hover:scale-125 ${newState.color === c ? 'ring-2 ring-primary ring-offset-2' : ''}`}
                                                        style={{ backgroundColor: c }}
                                                        onClick={() => setNewState({ ...newState, color: c })}
                                                    />
                                                ))}
                                            </div>
                                        </div>
                                    </div>

                                    <div className="flex justify-end gap-3 pt-4 border-t border-primary/10">
                                        <Button variant="ghost" size="sm" onClick={() => setIsAdding(false)}>
                                            放弃
                                        </Button>
                                        <Button size="sm" onClick={handleAdd} disabled={!newState.name} className="px-6">
                                            创建状态
                                        </Button>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
