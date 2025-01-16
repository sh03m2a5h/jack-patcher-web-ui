import type { NodeTypes } from '@xyflow/react';
import {AudioInputNode} from './AudioInputNode';
import {AudioOutputNode} from './AudioOutputNode';

import { AppNode } from './types';

export const initialNodes: AppNode[] = [
  {
    id: '561b7450-d26a-11ef-9049-6fa06980ea5c',
    type: 'audio-input',
    position: { x: 250, y: 100 },
    data: { deviceName: 'Microphone', channelCount: 1 },
  },
  {
    id: '7c1aa73e-d26a-11ef-97b4-1b72c5062fb7',
    type: 'audio-output',
    position: { x: 250, y: 300 },
    data: { deviceName: 'Speakers', channelCount: 2 },
  },
];

export const nodeTypes = {
  'audio-input': AudioInputNode,
  'audio-output': AudioOutputNode,
  // Add any of your custom nodes here!
} satisfies NodeTypes;
