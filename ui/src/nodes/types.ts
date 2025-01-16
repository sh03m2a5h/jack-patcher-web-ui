import type { Node } from '@xyflow/react';

export type AudioInputNode = Node<{ deviceName: string, channelCount: number }, 'audio-input'>;
export type AudioOutputNode = Node<{ deviceName: string, channelCount: number }, 'audio-output'>;
export type AppNode = AudioInputNode | AudioOutputNode;
