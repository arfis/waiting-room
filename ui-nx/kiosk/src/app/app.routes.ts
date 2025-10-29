import { Routes } from '@angular/router';
import { UserVerificationComponent } from './user-verification/user-verification.component';

export const routes: Routes = [
  { path: '', component: UserVerificationComponent },
  { path: '**', redirectTo: '' }
];
