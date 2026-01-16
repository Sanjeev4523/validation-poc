import type { ProtoFile } from '../types';

export const protoFiles: ProtoFile[] = [
  {
    name: 'Hello Request',
    description: 'Request message for greeting service',
    fullyQualifiedName: 'proto.HelloRequest',
  },
  {
    name: 'Hello Response',
    description: 'Response message from greeting service',
    fullyQualifiedName: 'proto.HelloResponse',
  },
  {
    name: 'Task',
    description: 'Task with name, description, timestamp, and status',
    fullyQualifiedName: 'proto.Task',
  },
  {
    name: 'UpdateTask',
    description: 'Request to update a task status. Comment is required when status is BLOCKED, otherwise optional',
    fullyQualifiedName: 'proto.UpdateTask',
  },
];
