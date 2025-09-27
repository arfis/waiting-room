import { Component, signal } from '@angular/core';
import { CardLoaderPageComponent } from './card-loader/card-loader.component';

@Component({
  selector: 'app-root',
  imports: [CardLoaderPageComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {
  protected readonly title = signal('kiosk');
}