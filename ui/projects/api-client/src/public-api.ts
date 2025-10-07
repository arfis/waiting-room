/*
 * Public API Surface of api-client
 */

export * from './lib/api-client';
export * from './lib/api';
export * from './lib/queue-websocket.service';

// Export types from api.ts
export type { 
  ConfigurationResponse, 
  RoomConfiguration, 
  ServicePointConfiguration,
  QueueEntry,
  QueueEntryStatus,
  PublicEntry,
  JoinResult,
  ApiConfig,
  ApiEnvironmentConfig
} from './lib/api';