import { ComponentFixture, TestBed } from '@angular/core/testing';
import { QueueList } from './queue-list';

describe('QueueList', () => {
  let component: QueueList;
  let fixture: ComponentFixture<QueueList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QueueList],
    }).compileComponents();

    fixture = TestBed.createComponent(QueueList);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
