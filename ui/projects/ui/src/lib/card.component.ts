import { Component, Input } from '@angular/core';
@Component({
  selector: 'ui-card',
  standalone: true,
  template: `<div class="rounded-2xl shadow p-4 bg-white"><ng-content/></div>`
})
export class CardComponent {}
