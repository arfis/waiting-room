import { Routes } from '@angular/router';
import { QueueComponent } from './queue/queue.component';

export const routes: Routes = [
  { path: '', redirectTo: '/q', pathMatch: 'full' },
  { path: 'q/:token', component: QueueComponent },
  { path: '**', redirectTo: '/q' }
];
