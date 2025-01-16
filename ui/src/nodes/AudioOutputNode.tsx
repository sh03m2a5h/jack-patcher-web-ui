import { Handle, Position, type NodeProps } from '@xyflow/react';
import { type AudioOutputNode } from './types';

export function AudioOutputNode({ data }: NodeProps<AudioOutputNode>) {
    return (
        <div className="react-flow__node-default flow-node">
            <div style={{ display: 'flex', alignItems: 'center' }}>
                <span>{data.deviceName}</span>
                <span className="material-symbols-outlined" style={{fontSize: 20}}>speaker</span>
            </div>
            {Array.from({ length: data.channelCount }, (_, i) => (
                <Handle key={i} type="target" position={Position.Left} id={`ch-${i}`} style={{
                    top: `${(i + 1) * 100 / (data.channelCount + 1)}%`,
                }} />
            ))}
        </div>
    );
}
