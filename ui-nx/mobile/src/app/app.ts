import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-root',
  imports: [CommonModule, RouterModule],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  protected readonly title = signal('mobile');
  queueData: any = null;

  constructor(private http: HttpClient) {}

  getCurrentPath(): string {
    return window.location.pathname;
  }

  getTokenFromPath(): string | null {
    const path = window.location.pathname;
    const match = path.match(/\/q\/(.+)/);
    return match ? match[1] : null;
  }

  loadQueueData() {
    const token = this.getTokenFromPath();
    if (token) {
      this.http.get(`http://localhost:8080/queue-entries/token/${token}`)
        .subscribe({
          next: (data) => {
            this.queueData = data;
            console.log('Queue data loaded:', data);
          },
          error: (error) => {
            console.error('Failed to load queue data:', error);
          }
        });
    }
  }
}
