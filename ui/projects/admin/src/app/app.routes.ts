import { Routes } from '@angular/router';
import { DashboardComponent } from './dashboard/dashboard';
import { ConfigurationComponent } from './configuration/configuration';
import { CardReadersComponent } from './card-readers/card-readers';

export const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  { path: 'dashboard', component: DashboardComponent },
  { path: 'configuration', component: ConfigurationComponent },
  { path: 'card-readers', component: CardReadersComponent },
  { path: '**', redirectTo: '/dashboard' }
];