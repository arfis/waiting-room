import { ComponentFixture, TestBed } from '@angular/core/testing';
import { QueueItem } from './queue-item';

describe('QueueItem', () => {
  let component: QueueItem;
  let fixture: ComponentFixture<QueueItem>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QueueItem],
    }).compileComponents();

    fixture = TestBed.createComponent(QueueItem);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
