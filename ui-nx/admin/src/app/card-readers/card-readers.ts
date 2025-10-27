import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

interface CardReader {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'error';
  ipAddress?: string;
  version?: string;
  lastSeen: string;
  lastError?: string;
}

@Component({
  selector: 'app-card-readers',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './card-readers.html',
  styleUrl: './card-readers.scss'
})
export class CardReadersComponent {
  isRefreshing = signal(false);
  selectedReader = signal<CardReader | null>(null);
  restartingReaders = signal<Set<string>>(new Set());
  
  cardReaders = signal<CardReader[]>([
    {
      id: 'cr-001',
      name: 'Kiosk Card Reader 1',
      status: 'online',
      ipAddress: '192.168.1.100',
      version: '1.2.3',
      lastSeen: new Date().toISOString()
    },
    {
      id: 'cr-002',
      name: 'Kiosk Card Reader 2',
      status: 'offline',
      ipAddress: '192.168.1.101',
      version: '1.2.2',
      lastSeen: new Date(Date.now() - 300000).toISOString() // 5 minutes ago
    },
    {
      id: 'cr-003',
      name: 'Kiosk Card Reader 3',
      status: 'error',
      ipAddress: '192.168.1.102',
      version: '1.2.1',
      lastSeen: new Date(Date.now() - 600000).toISOString(), // 10 minutes ago
      lastError: 'Connection timeout'
    }
  ]);

  onlineCount = signal(1);
  offlineCount = signal(1);
  errorCount = signal(1);

  constructor() {
    this.updateCounts();
  }

  refreshStatus(): void {
    this.isRefreshing.set(true);
    
    // Simulate API call
    setTimeout(() => {
      this.isRefreshing.set(false);
      this.updateCounts();
      console.log('Card reader status refreshed');
    }, 1000);
  }

  restartCardReader(readerId: string): void {
    this.restartingReaders.update(set => new Set([...set, readerId]));
    
    // Simulate restart API call
    setTimeout(() => {
      this.restartingReaders.update(set => {
        const newSet = new Set(set);
        newSet.delete(readerId);
        return newSet;
      });
      
      // Update reader status to online after restart
      this.cardReaders.update(readers => 
        readers.map(reader => 
          reader.id === readerId 
            ? { ...reader, status: 'online' as const, lastSeen: new Date().toISOString() }
            : reader
        )
      );
      
      this.updateCounts();
      console.log(`Card reader ${readerId} restarted`);
    }, 2000);
  }

  viewDetails(reader: CardReader): void {
    this.selectedReader.set(reader);
  }

  closeDetails(): void {
    this.selectedReader.set(null);
  }

  isRestarting(readerId: string): boolean {
    return this.restartingReaders().has(readerId);
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'online':
        return 'bg-green-100 text-green-800';
      case 'offline':
        return 'bg-red-100 text-red-800';
      case 'error':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }

  formatLastSeen(lastSeen: string | undefined): string {
    if (!lastSeen) return 'Never';
    const date = new Date(lastSeen);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) {
      return 'Just now';
    } else if (diffMins < 60) {
      return `${diffMins} minutes ago`;
    } else if (diffMins < 1440) {
      const hours = Math.floor(diffMins / 60);
      return `${hours} hours ago`;
    } else {
      return date.toLocaleDateString();
    }
  }

  private updateCounts(): void {
    const readers = this.cardReaders();
    this.onlineCount.set(readers.filter(r => r.status === 'online').length);
    this.offlineCount.set(readers.filter(r => r.status === 'offline').length);
    this.errorCount.set(readers.filter(r => r.status === 'error').length);
  }
}