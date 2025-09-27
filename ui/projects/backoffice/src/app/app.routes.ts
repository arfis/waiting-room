import { Routes } from '@angular/router';
import { QueueManagementComponent } from './queue-management/queue-management.component';

export const routes: Routes = [
  { path: '', component: QueueManagementComponent },
  { path: '**', redirectTo: '' }
];
