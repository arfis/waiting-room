// Just for demo purposes when the clients dont exist yet
export interface JoinResult { entryId: string; ticketNumber: string; qrUrl: string; }
export type QueueEntryStatus = 'WAITING'|'CALLED'|'IN_SERVICE'|'COMPLETED'|'SKIPPED'|'CANCELLED'|'NO_SHOW';
export interface PublicEntry { entryId: string; ticketNumber: string; status: QueueEntryStatus; position: number; etaMinutes: number; canCancel: boolean; }
export interface QueueEntry { id: string; waitingRoomId: string; ticketNumber: string; status: QueueEntryStatus; position: number; }

export interface ApiConfig { baseUrl?: string; }
export class Api {
  constructor(private cfg: ApiConfig = {}) {}
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
