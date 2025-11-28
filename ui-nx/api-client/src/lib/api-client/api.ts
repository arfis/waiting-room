// Just for demo purposes when the clients dont exist yet
export interface JoinResult { entryId: string; ticketNumber: string; qrUrl: string; }
export type QueueEntryStatus = 'WAITING'|'CALLED'|'IN_SERVICE'|'COMPLETED'|'SKIPPED'|'CANCELLED'|'NO_SHOW';
export interface PublicEntry { entryId: string; ticketNumber: string; status: QueueEntryStatus; position: number; etaMinutes: number; canCancel: boolean; servicePoint?: string; }
export interface QueueEntry { id: string; waitingRoomId: string; ticketNumber: string; status: QueueEntryStatus; position: number; servicePoint?: string; serviceName?: string; serviceDuration?: number; age?: number; symbols?: string[]; }

export interface ServicePointConfiguration { id: string; name: string; description?: string; managerId?: string; managerName?: string; }
export interface RoomConfiguration { id: string; name: string; servicePoints: ServicePointConfiguration[]; }
export interface ConfigurationResponse {
  defaultRoom: string;
  allowWildcard: boolean;
  webSocketPath: string;
  rooms: RoomConfiguration[];
}
export interface ApiEnvironmentConfig {
  apiUrl?: string;
}

export interface ApiConfig { baseUrl?: string; }

export const DEFAULT_API_BASE_URL = 'http://localhost:8080/api';

export class Api {
  private readonly baseUrl: string;
  constructor(private cfg: ApiConfig = {}) {
    const base = cfg.baseUrl ?? DEFAULT_API_BASE_URL;
    this.baseUrl = base.endsWith('/') ? base.slice(0, -1) : base;
  }

  configuration = {
    getConfiguration: async (): Promise<ConfigurationResponse> => {
      const response = await fetch(`${this.baseUrl}/config`, { credentials: 'include' });
      if (!response.ok) {
        throw new Error(`Failed to load configuration (status ${response.status})`);
      }
      return response.json() as Promise<ConfigurationResponse>;
    }
  }

  kiosk = {
    postWaitingRoomsRoomIdSwipe: async (roomId: string, body: { idCardRaw: string }): Promise<JoinResult> => {
      await wait(300);
      return {
        entryId: cryptoId(),
        ticketNumber: `A-${Math.floor(Math.random()*900+100)}`,
        qrUrl: `http://localhost:4202/q/${cryptoId(8)}`
      };
    }
  }
  queue = {
    getQueueEntriesTokenQrToken: async (token: string): Promise<PublicEntry> => {
      await wait(200);
      const pos = Math.floor(Math.random()*10)+1;
      return {
        entryId: cryptoId(),
        ticketNumber: `A-${Math.floor(Math.random()*900+100)}`,
        status: pos === 1 ? 'CALLED' : 'WAITING',
        position: pos,
        etaMinutes: pos*3,
        canCancel: false,
      }
    },
    getWaitingRoomsRoomIdQueue: async (roomId: string, state?: QueueEntryStatus): Promise<QueueEntry[]> => {
      await wait(200);
      const entries: QueueEntry[] = [];
      const count = Math.floor(Math.random()*8)+2; // 2-9 entries
      
      for (let i = 0; i < count; i++) {
        const entryState = state || (Math.random() > 0.7 ? 'CALLED' : 'WAITING');
        entries.push({
          id: cryptoId(),
          waitingRoomId: roomId,
          ticketNumber: `A-${Math.floor(Math.random()*900+100)}`,
          status: entryState,
          position: i + 1,
        });
      }
      
      // Filter by state if provided
      return state ? entries.filter(entry => entry.status === state) : entries;
    }
  }
  waitingRoom = {
    postWaitingRoomsRoomIdNext: async (roomId: string): Promise<QueueEntry> => {
      await wait(200);
      return {
        id: cryptoId(),
        waitingRoomId: roomId,
        ticketNumber: `A-${Math.floor(Math.random()*900+100)}`,
        status: 'CALLED',
        position: 0,
      }
    }
  }
}
function wait(ms:number){ return new Promise(r=>setTimeout(r,ms)); }
function cryptoId(len=36){ const s=crypto.getRandomValues(new Uint8Array(len)).reduce((a,b)=>a+((b%36).toString(36)),''); return s.slice(0,len); }

export function createApiClient(environment?: ApiEnvironmentConfig, overrides: ApiConfig = {}): Api {
  const baseUrl = overrides.baseUrl ?? environment?.apiUrl ?? DEFAULT_API_BASE_URL;
  return new Api({ ...overrides, baseUrl });
}
