import type { TeamMember } from '../../types';
import { Avatar, Badge, MoreHorizontalIcon, ChevronRightIcon } from '../atoms';
import { cn } from '../../utils';

export interface AdminMembersTableProps {
  members: TeamMember[];
  className?: string;
}

export const AdminMembersTable = ({ members, className }: AdminMembersTableProps) => {
  return (
    <div
      className={cn(
        'bg-white rounded-2xl border border-border p-6 shadow-soft-sm space-y-4',
        className
      )}
    >
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 pb-2">
        <div>
          <h4 className="text-base font-bold text-text-primary tracking-tight">
            Team members
          </h4>
          <p className="text-xs text-text-secondary font-medium">
            Manage people and workspace access
          </p>
        </div>

        <button
          type="button"
          className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:text-primary-hover transition-colors cursor-pointer self-start sm:self-auto"
        >
          <span>View all</span>
          <ChevronRightIcon className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Table Container */}
      <div className="overflow-x-auto -mx-6 px-6">
        <table className="w-full text-left border-collapse min-w-[600px]">
          <thead>
            <tr className="border-b border-border text-[11px] font-bold text-text-secondary uppercase tracking-wider">
              <th className="py-3 px-3">Member</th>
              <th className="py-3 px-3">Role</th>
              <th className="py-3 px-3">Tasks Completed</th>
              <th className="py-3 px-3">Status</th>
              <th className="py-3 px-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-subtle text-xs">
            {members.map((member) => (
              <tr
                key={member.id}
                className="hover:bg-surface-subtle transition-colors group"
              >
                {/* Member Info */}
                <td className="py-3.5 px-3">
                  <div className="flex items-center gap-3">
                    <Avatar initials={member.avatarText} size="sm" />
                    <div>
                      <p className="font-semibold text-text-primary group-hover:text-primary transition-colors">
                        {member.name}
                      </p>
                      <p className="text-[11px] text-text-secondary">
                        {member.email}
                      </p>
                    </div>
                  </div>
                </td>

                {/* Role */}
                <td className="py-3.5 px-3 font-medium text-text-primary">
                  {member.role}
                </td>

                {/* Aggregated Tasks Completed (Privacy Compliant) */}
                <td className="py-3.5 px-3 font-semibold text-text-primary">
                  {member.tasksCompleted}
                </td>

                {/* Status Badge */}
                <td className="py-3.5 px-3">
                  <Badge
                    variant={
                      member.status === 'Active'
                        ? 'emerald'
                        : member.status === 'Away'
                        ? 'amber'
                        : 'slate'
                    }
                  >
                    <span
                      className={cn(
                        'w-1.5 h-1.5 rounded-full mr-1',
                        member.status === 'Active'
                          ? 'bg-emerald-500'
                          : member.status === 'Away'
                          ? 'bg-amber-500'
                          : 'bg-slate-400'
                      )}
                    />
                    {member.status}
                  </Badge>
                </td>

                {/* Action button */}
                <td className="py-3.5 px-3 text-right">
                  <button
                    type="button"
                    className="p-1.5 text-text-secondary hover:text-text-primary hover:bg-surface-muted rounded-lg transition-colors cursor-pointer"
                    aria-label={`Actions for ${member.name}`}
                  >
                    <MoreHorizontalIcon className="w-4 h-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
