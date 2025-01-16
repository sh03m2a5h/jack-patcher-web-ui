import { Handle, Position, type NodeProps } from '@xyflow/react';
import { type AudioInputNode } from './types';

export function AudioInputNode({ data }: NodeProps<AudioInputNode>) {
    return (
        <div className="react-flow__node-default flow-node">
        <div style={{ display: 'flex', alignItems: 'center' }}>
            <span className="material-symbols-outlined" style={{fontSize: 20}}>mic</span>
            <span>{data.deviceName}</span>
        </div>
            {Array.from({ length: data.channelCount }, (_, i) => (
                <Handle key={i} type="source" position={Position.Right} id={`ch-${i}`} style={{
                    top: `${(i + 1) * 100 / (data.channelCount + 1)}%`,
                }} />
            ))}
        </div>
    );
}
