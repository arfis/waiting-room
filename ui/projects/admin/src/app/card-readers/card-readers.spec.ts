import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CardReaders } from './card-readers';

describe('CardReaders', () => {
  let component: CardReaders;
  let fixture: ComponentFixture<CardReaders>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CardReaders]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CardReaders);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
