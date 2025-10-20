import { Routes } from '@angular/router';
import { CheckInEntryComponent } from './check-in-entry/check-in-entry.component';
import { CardLoaderComponent } from './card-loader/card-loader.component';
import { ManualIdComponent } from './manual/manual.component';

export const routes: Routes = [
  { path: '', component: CheckInEntryComponent },
  { path: 'card', component: CardLoaderComponent },
  { path: 'manual', component: ManualIdComponent },
  { path: '**', redirectTo: '' }
];
