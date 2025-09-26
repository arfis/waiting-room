import { Component, signal } from '@angular/core';
import { KioskComponent } from './kiosk.component';

@Component({
  selector: 'app-root',
  imports: [KioskComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  protected readonly title = signal('kiosk');
}
