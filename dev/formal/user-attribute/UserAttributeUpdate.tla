---- MODULE UserAttributeUpdate ----
(*
User values of one lightdash_user_attribute during a Terraform update.

Update (internal/lightdash/services/user_attributes.go) reads the attribute, then writes
the whole attribute back with the user values that it read. The API replaces the whole
attribute, so any user value that changes between the read and the write is lost.

  Get: UserAttributesService.Update calls GetByUUID, which lists the attributes.
  Put: keepUserValues copies the read users into the request, then
       apiv1.UpdateUserAttributeV1 sends it.
  UIEdit: a user changes the value in the Lightdash UI, outside Terraform.

KeepUsers = FALSE models a regression where the PUT sends no user values.
EditsDuringApply = FALSE models a UI that is only edited between applies.
*)
EXTENDS Naturals, TLC

CONSTANTS
  UserValues,         \* values that a user can hold in the attribute
  Empty,              \* the value after a PUT that sends no user values
  KeepUsers,          \* TRUE: the PUT sends back the users that Get read
  EditsDuringApply,   \* TRUE: a user can change the value while Update runs
  MaxEdits            \* bound on the UI edits in one run

ASSUME Empty \notin UserValues
ASSUME KeepUsers \in BOOLEAN
ASSUME EditsDuringApply \in BOOLEAN
ASSUME MaxEdits \in Nat

VARIABLES
  serverUsers,  \* the user values that Lightdash stores
  lastUI,       \* the value that the last UI edit (or the start) set
  snapshot,     \* the user values that Get read
  pc,           \* "idle" -> "read" -> "done"
  edits         \* UI edits so far

vars == <<serverUsers, lastUI, snapshot, pc, edits>>

Init ==
  /\ serverUsers \in UserValues
  /\ lastUI = serverUsers
  /\ snapshot = Empty
  /\ pc = "idle"
  /\ edits = 0

\* A user edits the value in the Lightdash UI, outside Terraform.
UIEdit ==
  /\ edits < MaxEdits
  /\ \/ pc = "idle"
     \/ EditsDuringApply /\ pc = "read"
  /\ \E v \in UserValues :
       /\ serverUsers' = v
       /\ lastUI' = v
  /\ edits' = edits + 1
  /\ UNCHANGED <<snapshot, pc>>

\* Update, first request: services Update calls GetByUUID.
Get ==
  /\ pc = "idle"
  /\ snapshot' = serverUsers
  /\ pc' = "read"
  /\ UNCHANGED <<serverUsers, lastUI, edits>>

\* Update, second request: keepUserValues, then apiv1.UpdateUserAttributeV1.
Put ==
  /\ pc = "read"
  /\ serverUsers' = IF KeepUsers THEN snapshot ELSE Empty
  /\ pc' = "done"
  /\ UNCHANGED <<lastUI, snapshot, edits>>

Done == pc = "done" /\ UNCHANGED vars

Next == UIEdit \/ Get \/ Put \/ Done

Spec == Init /\ [][Next]_vars

\* After an apply, the value that a user last set in Lightdash is still stored.
ValueSurvivesApply == pc = "done" => serverUsers = lastUI
====
