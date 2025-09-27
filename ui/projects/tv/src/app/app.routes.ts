import { Routes } from '@angular/router';
import { QueueDisplayComponent } from './queue-display/queue-display.component';

export const routes: Routes = [
  { path: '', component: QueueDisplayComponent },
  { path: '**', redirectTo: '' }
];
