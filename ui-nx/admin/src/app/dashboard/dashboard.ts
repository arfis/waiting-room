import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss'
})
export class DashboardComponent {
  // Mock data - in real implementation, these would come from services
  activeRooms = signal(3);
  onlineCardReaders = signal(2);
  totalCardReaders = signal(3);
  waitingCount = signal(12);
  systemStatus = signal('Healthy');
}