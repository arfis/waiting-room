import { Component } from '@angular/core';
import { CardLoaderComponent } from './card-loader/card-loader.component';

@Component({
  selector: 'app-root',
  imports: [CardLoaderComponent],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class AppComponent {
  protected readonly title = 'kiosk';
}