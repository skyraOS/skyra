/-! Proposal 58: shapes. The shape in a base type is inferred by the core
from what arrives, never supplied by the shell. A shape is the keys,
recursively, with each leaf's JSON kind. Key order, array order and array
length are not in it; key sets are strict; there is one number kind;
duplicate keys have no shape.

What arrives is decided by its contract's declared output: under `text` it
is a string whatever it looks like; under `json` it is decoded. Output that
does not decode arrives as an error, `{error: string}`, so every arrival
has a shape.

The canonical encoding is what the base's hash is taken over. Lean never
computes a hash; that the encoding is injective is assumed where it is
needed, as `Hashes.Sound` assumes the hashes are. -/

namespace Shape

/-! ## Values and shapes -/

/-- A JSON value, as decoded: an object's members in the order written,
duplicates included, so that a duplicate can be refused. A number's text
is kept; its value is never in a shape. -/
inductive Value where
  | null
  | bool (b : Bool)
  | num (lit : String)
  | str (s : String)
  | arr (xs : List Value)
  | obj (kvs : List (String × Value))
  /-- Untrusted bytes that did not decode. JSON has no such kind: only the
  system makes it. -/
  | bytes (b : ByteArray)
  /-- A frame's exit status. JSON has no such kind: only a frame's return
  makes it. -/
  | exit (status : Nat)

inductive Shape where
  | null
  | bool
  | num
  | str
  | bytes
  /-- An exit status, the status in the shape: two statuses are two shapes,
  so they are two bases. -/
  | exit (status : Nat)
  /-- The shapes of its elements, as a set: sorted by encoding, once each. -/
  | arr (elems : List Shape)
  /-- Its keys, each with its value's shape, sorted by key. -/
  | obj (fields : List (String × Shape))
  deriving Repr

/-! ## The canonical encoding -/

def hexDigit (n : Nat) : Char :=
  if n < 10 then Char.ofNat (48 + n) else Char.ofNat (87 + n)

/-- A key as a JSON string's contents: `"` and `\` escaped, and control
characters as `\u00xx`. -/
def escapeChar (c : Char) : String :=
  if c = '"' then "\\\""
  else if c = '\\' then "\\\\"
  else if c.toNat < 32 then "\\u00" ++ (hexDigit (c.toNat / 16)).toString ++ (hexDigit (c.toNat % 16)).toString
  else c.toString

def escape (s : String) : String := s.foldl (fun acc c => acc ++ escapeChar c) ""

mutual
def Shape.enc : Shape → String
  | .null => "null"
  | .bool => "boolean"
  | .num => "number"
  | .str => "string"
  | .bytes => "bytes"
  | .exit n => "exit(" ++ toString n ++ ")"
  | .arr es => "[" ++ encElems es ++ "]"
  | .obj fs => "{" ++ encFields fs ++ "}"

def encElems : List Shape → String
  | [] => ""
  | [e] => e.enc
  | e :: es => e.enc ++ "," ++ encElems es

def encFields : List (String × Shape) → String
  | [] => ""
  | [(k, s)] => "\"" ++ escape k ++ "\":" ++ s.enc
  | (k, s) :: fs => "\"" ++ escape k ++ "\":" ++ s.enc ++ "," ++ encFields fs
end

/-- What is assumed of the encoding: two shapes encode alike only if they
are one. -/
def EncInjective : Prop := ∀ a b : Shape, a.enc = b.enc → a = b

/-! ## Sorting and uniqueness -/

/-- Two permutations both ordered by an asymmetric relation are equal. -/
theorem eq_of_perm_of_pairwise {R : α → α → Prop} (asymm : ∀ a b, R a b → ¬ R b a) :
    ∀ {l₁ l₂ : List α}, l₁.Perm l₂ → l₁.Pairwise R → l₂.Pairwise R → l₁ = l₂
  | [], _, h, _, _ => h.nil_eq
  | a :: as, [], h, _, _ => absurd h.symm.nil_eq (by simp)
  | a :: as, b :: bs, h, p₁, p₂ => by
    rw [List.pairwise_cons] at p₁ p₂
    have hab : a = b := by
      by_cases e : a = b
      · exact e
      · have ha : a ∈ bs := by
          have := h.subset (List.mem_cons_self ..)
          rcases List.mem_cons.mp this with h' | h'
          · exact absurd h' e
          · exact h'
        have hb : b ∈ as := by
          have := h.symm.subset (List.mem_cons_self ..)
          rcases List.mem_cons.mp this with h' | h'
          · exact absurd h'.symm e
          · exact h'
        exact absurd (p₁.1 b hb) (asymm _ _ (p₂.1 a ha))
    subst hab
    rw [eq_of_perm_of_pairwise asymm (List.perm_cons a |>.mp h) p₁.2 p₂.2]

def keyLe (a b : String × Shape) : Bool := decide (a.1 ≤ b.1)

def sortKeys (fs : List (String × Shape)) : List (String × Shape) := fs.mergeSort keyLe

def encLe (a b : Shape) : Bool := decide (a.enc ≤ b.enc)

/-- Once each by `f`, keeping the last of each. -/
def dedupBy (f : α → String) : List α → List α
  | [] => []
  | x :: xs => if xs.any (f · == f x) then dedupBy f xs else x :: dedupBy f xs

theorem mem_dedupBy {f : α → String} : ∀ {l : List α} {a}, a ∈ dedupBy f l → a ∈ l
  | [], _, h => by simp [dedupBy] at h
  | x :: xs, a, h => by
    unfold dedupBy at h; split at h
    · exact List.mem_cons_of_mem _ (mem_dedupBy h)
    · rcases List.mem_cons.mp h with rfl | h
      · simp
      · exact List.mem_cons_of_mem _ (mem_dedupBy h)

theorem dedupBy_covers {f : α → String} : ∀ {l : List α} {a}, a ∈ l → ∃ b ∈ dedupBy f l, f b = f a
  | [], _, h => by simp at h
  | x :: xs, a, h => by
    unfold dedupBy; split
    · rename_i hany
      rcases List.mem_cons.mp h with rfl | h
      · obtain ⟨y, hy, e⟩ := List.any_eq_true.mp hany
        obtain ⟨b, hb, e'⟩ := dedupBy_covers (f := f) hy
        exact ⟨b, hb, e'.trans (by simpa using e)⟩
      · exact dedupBy_covers h
    · rcases List.mem_cons.mp h with rfl | h
      · exact ⟨_, List.mem_cons_self .., rfl⟩
      · obtain ⟨b, hb, e⟩ := dedupBy_covers (f := f) h
        exact ⟨b, List.mem_cons_of_mem _ hb, e⟩

theorem nodup_dedupBy {f : α → String} : ∀ {l : List α}, ((dedupBy f l).map f).Nodup
  | [] => by simp [dedupBy]
  | x :: xs => by
    unfold dedupBy; split
    · exact nodup_dedupBy
    · rename_i hany
      refine List.nodup_cons.mpr ⟨fun hm => hany ?_, nodup_dedupBy⟩
      obtain ⟨y, hy, e⟩ := List.mem_map.mp hm
      exact List.any_eq_true.mpr ⟨y, mem_dedupBy hy, by simp [e]⟩

theorem nodup_of_map {f : α → β} : ∀ {l : List α}, (l.map f).Nodup → l.Nodup
  | [], _ => List.nodup_nil
  | x :: xs, h => by
    rw [List.map_cons, List.nodup_cons] at h
    exact List.nodup_cons.mpr ⟨fun hx => h.1 (List.mem_map_of_mem hx), nodup_of_map h.2⟩

def canonArr (ss : List Shape) : Shape := .arr ((dedupBy Shape.enc ss).mergeSort encLe)

/-- Ordered by `≤` and never equal: strictly ordered, so a permutation
fixes the order. -/
theorem pairwise_strict {f : α → String} {le : α → α → Bool}
    (hle : ∀ a b, le a b = decide (f a ≤ f b)) {l : List α}
    (hs : l.Pairwise (le · ·)) (hn : (l.map f).Nodup) :
    l.Pairwise (fun a b => f a ≤ f b ∧ f a ≠ f b) := by
  have := hs.and (List.pairwise_map.mp hn)
  exact this.imp fun ⟨h₁, h₂⟩ => ⟨by simpa [hle] using h₁, h₂⟩

theorem strict_asymm {f : α → String} :
    ∀ a b, (f a ≤ f b ∧ f a ≠ f b) → ¬ (f b ≤ f a ∧ f b ≠ f a) :=
  fun _ _ ⟨h₁, h₂⟩ ⟨h₃, _⟩ => h₂ (String.le_antisymm h₁ h₃)

theorem sorted_by {f : α → String} {le : α → α → Bool}
    (hle : ∀ a b, le a b = decide (f a ≤ f b)) (l : List α) : (l.mergeSort le).Pairwise (le · ·) :=
  List.pairwise_mergeSort (fun a b c h₁ h₂ => by simp only [hle, decide_eq_true_eq] at *; exact String.le_trans h₁ h₂)
    (fun a b => by simp only [hle]; rcases String.le_total (f a) (f b) with h | h <;> simp [h]) l

/-- Two lists with the same keys, each once, sort to one list. -/
theorem sort_unique {f : α → String} {le : α → α → Bool}
    (hle : ∀ a b, le a b = decide (f a ≤ f b)) {l₁ l₂ : List α} (h : l₁.Perm l₂)
    (hn : (l₁.map f).Nodup) : l₁.mergeSort le = l₂.mergeSort le := by
  have p₁ := List.mergeSort_perm l₁ le
  have p₂ := List.mergeSort_perm l₂ le
  have n₁ : ((l₁.mergeSort le).map f).Nodup := (p₁.map f).nodup_iff.mpr hn
  have n₂ : ((l₂.mergeSort le).map f).Nodup := (p₂.map f).nodup_iff.mpr ((h.map f).nodup_iff.mp hn)
  exact eq_of_perm_of_pairwise strict_asymm (p₁.trans (h.trans p₂.symm))
    (pairwise_strict hle (sorted_by hle l₁) n₁) (pairwise_strict hle (sorted_by hle l₂) n₂)

/-- An array's shape is the set of its elements' shapes: order and
repetition do not matter. -/
theorem canonArr_eq (inj : EncInjective) {ss ts : List Shape} (h : ∀ s, s ∈ ss ↔ s ∈ ts) :
    canonArr ss = canonArr ts := by
  unfold canonArr
  congr 1
  have mem : ∀ {l : List Shape} {s}, s ∈ dedupBy Shape.enc l ↔ s ∈ l := fun {l s} =>
    ⟨mem_dedupBy, fun hs => by
      obtain ⟨b, hb, e⟩ := dedupBy_covers (f := Shape.enc) hs
      exact inj _ _ e ▸ hb⟩
  have nd : ∀ {l : List Shape}, (dedupBy Shape.enc l).Nodup := fun {l} =>
    nodup_of_map nodup_dedupBy
  have perm : (dedupBy Shape.enc ss).Perm (dedupBy Shape.enc ts) :=
    (List.perm_ext_iff_of_nodup nd nd).mpr fun s => by rw [mem, mem, h]
  exact sort_unique (f := Shape.enc) (fun _ _ => rfl) perm nodup_dedupBy

/-! ## Inference -/

mutual
/-- A value's shape, or none when an object repeats a key, at any depth. -/
def shapeOf : Value → Option Shape
  | .null => some .null
  | .bool _ => some .bool
  | .num _ => some .num
  | .str _ => some .str
  | .arr xs => (shapesOf xs).map canonArr
  | .obj kvs => if (kvs.map (·.1)).Nodup then (fieldsOf kvs).map (.obj ∘ sortKeys) else none
  | .bytes _ => some .bytes
  | .exit n => some (.exit n)

def shapesOf : List Value → Option (List Shape)
  | [] => some []
  | x :: xs => match shapeOf x, shapesOf xs with
    | some s, some ss => some (s :: ss)
    | _, _ => none

def fieldsOf : List (String × Value) → Option (List (String × Shape))
  | [] => some []
  | (k, v) :: kvs => match shapeOf v, fieldsOf kvs with
    | some s, some fs => some ((k, s) :: fs)
    | _, _ => none
end

/-! ## What inference keeps and forgets -/

theorem shapesOf_mem : ∀ {xs : List Value} {ss}, shapesOf xs = some ss →
    ∀ s, s ∈ ss ↔ ∃ x ∈ xs, shapeOf x = some s
  | [], ss, h, s => by simp [shapesOf] at h; subst h; simp
  | x :: xs, ss, h, s => by
    unfold shapesOf at h
    split at h
    · rename_i s₀ ss₀ hx hxs
      cases h
      rw [List.mem_cons, shapesOf_mem hxs s]
      constructor
      · rintro (rfl | ⟨y, hy, e⟩)
        · exact ⟨x, List.mem_cons_self .., hx⟩
        · exact ⟨y, List.mem_cons_of_mem x hy, e⟩
      · rintro ⟨y, hy, e⟩
        rcases List.mem_cons.mp hy with rfl | hy
        · exact .inl (Option.some.inj (e.symm.trans hx))
        · exact .inr ⟨y, hy, e⟩
    · cases h

theorem shapesOf_none : ∀ {xs : List Value}, shapesOf xs = none → ∃ x ∈ xs, shapeOf x = none
  | [], h => by simp [shapesOf] at h
  | x :: xs, h => by
    unfold shapesOf at h
    split at h
    · cases h
    · rename_i hne
      cases hx : shapeOf x with
      | none => exact ⟨x, List.mem_cons_self .., hx⟩
      | some s =>
        cases hxs : shapesOf xs with
        | none => obtain ⟨y, hy, e⟩ := shapesOf_none hxs; exact ⟨y, List.mem_cons_of_mem _ hy, e⟩
        | some ss => exact absurd (hne s ss hx hxs) id

theorem shapesOf_all : ∀ {xs : List Value} {ss}, shapesOf xs = some ss → ∀ x ∈ xs, (shapeOf x).isSome
  | [], _, _, x, hx => by simp at hx
  | x :: xs, ss, h, y, hy => by
    unfold shapesOf at h
    split at h
    · rename_i s₀ ss₀ hx hxs
      rcases List.mem_cons.mp hy with rfl | hy
      · simp [hx]
      · exact shapesOf_all hxs y hy
    · cases h

/-- Array order and length are values, not shape: two arrays whose
elements have the same shapes, as sets, have one shape. -/
theorem shapeOf_arr (inj : EncInjective) {xs ys : List Value}
    (hx : ∀ x ∈ xs, ∃ y ∈ ys, shapeOf x = shapeOf y)
    (hy : ∀ y ∈ ys, ∃ x ∈ xs, shapeOf y = shapeOf x) :
    shapeOf (.arr xs) = shapeOf (.arr ys) := by
  simp only [shapeOf]
  cases h₁ : shapesOf xs with
  | none =>
    cases h₂ : shapesOf ys with
    | none => rfl
    | some ts =>
      obtain ⟨x, hxm, e⟩ := shapesOf_none h₁
      obtain ⟨y, hym, e'⟩ := hx x hxm
      have := shapesOf_all h₂ y hym
      rw [← e', e] at this; cases this
  | some ss =>
    cases h₂ : shapesOf ys with
    | none =>
      obtain ⟨y, hym, e⟩ := shapesOf_none h₂
      obtain ⟨x, hxm, e'⟩ := hy y hym
      have := shapesOf_all h₁ x hxm
      rw [← e', e] at this; cases this
    | some ts =>
      simp only [Option.map_some, Option.some.injEq]
      apply canonArr_eq inj
      intro s
      rw [shapesOf_mem h₁, shapesOf_mem h₂]
      constructor
      · rintro ⟨x, hxm, e⟩; obtain ⟨y, hym, e'⟩ := hx x hxm; exact ⟨y, hym, e' ▸ e⟩
      · rintro ⟨y, hym, e⟩; obtain ⟨x, hxm, e'⟩ := hy y hym; exact ⟨x, hxm, e' ▸ e⟩

theorem inj_of_nodup_map {f : α → β} :
    ∀ {l : List α}, (l.map f).Nodup → ∀ {x y}, x ∈ l → y ∈ l → f x = f y → x = y
  | [], _, _, _, hx, _, _ => by simp at hx
  | a :: l, h, x, y, hx, hy, e => by
    rw [List.map_cons, List.nodup_cons] at h
    rcases List.mem_cons.mp hx with rfl | hx' <;> rcases List.mem_cons.mp hy with rfl | hy'
    · rfl
    · exact absurd (e ▸ List.mem_map_of_mem hy') h.1
    · exact absurd (e.symm ▸ List.mem_map_of_mem hx') h.1
    · exact inj_of_nodup_map h.2 hx' hy' e

theorem sortKeys_perm {fs : List (String × Shape)} : (sortKeys fs).Perm fs :=
  List.mergeSort_perm fs keyLe

theorem fieldsOf_cons {k : String} {v : Value} {kvs fs} : fieldsOf ((k, v) :: kvs) = some fs ↔
    ∃ s fs₀, shapeOf v = some s ∧ fieldsOf kvs = some fs₀ ∧ fs = (k, s) :: fs₀ := by
  constructor
  · intro h; unfold fieldsOf at h; split at h
    · rename_i s fs₀ hv hk; cases h; exact ⟨s, fs₀, hv, hk, rfl⟩
    · cases h
  · rintro ⟨s, fs₀, hv, hk, rfl⟩; simp [fieldsOf, hv, hk]

theorem fieldsOf_keys : ∀ {kvs : List (String × Value)} {fs}, fieldsOf kvs = some fs →
    fs.map (·.1) = kvs.map (·.1)
  | [], fs, h => by simp [fieldsOf] at h; subst h; rfl
  | (k, v) :: kvs, fs, h => by
    obtain ⟨s, fs₀, -, hk, rfl⟩ := fieldsOf_cons.mp h
    simp [fieldsOf_keys hk]

theorem fieldsOf_mem : ∀ {kvs : List (String × Value)} {fs}, fieldsOf kvs = some fs →
    ∀ {k v}, (k, v) ∈ kvs → ∃ s, shapeOf v = some s ∧ (k, s) ∈ fs
  | [], _, _, _, _, h => by simp at h
  | (k', v') :: kvs, fs, h, k, v, hm => by
    obtain ⟨s, fs₀, hv, hk, rfl⟩ := fieldsOf_cons.mp h
    rcases List.mem_cons.mp hm with e | hm
    · simp only [Prod.mk.injEq] at e; obtain ⟨rfl, rfl⟩ := e
      exact ⟨s, hv, List.mem_cons_self ..⟩
    · obtain ⟨t, ht, hmt⟩ := fieldsOf_mem hk hm; exact ⟨t, ht, List.mem_cons_of_mem _ hmt⟩

theorem fieldsOf_perm {kvs kvs' : List (String × Value)} (h : kvs.Perm kvs') :
    ∀ {fs}, fieldsOf kvs = some fs → ∃ fs', fieldsOf kvs' = some fs' ∧ fs.Perm fs' := by
  induction h with
  | nil => intro fs h; exact ⟨fs, h, .refl _⟩
  | cons x _ ih =>
    intro fs hf
    obtain ⟨k, v⟩ := x
    obtain ⟨s, fs₀, hv, hk, rfl⟩ := fieldsOf_cons.mp hf
    obtain ⟨fs₀', h', p⟩ := ih hk
    exact ⟨_, fieldsOf_cons.mpr ⟨s, fs₀', hv, h', rfl⟩, p.cons _⟩
  | swap x y l =>
    intro fs hf
    obtain ⟨k, v⟩ := x; obtain ⟨k', v'⟩ := y
    obtain ⟨s', fs₁, hv', h₁, rfl⟩ := fieldsOf_cons.mp hf
    obtain ⟨s, fs₀, hv, h₀, rfl⟩ := fieldsOf_cons.mp h₁
    exact ⟨_, fieldsOf_cons.mpr ⟨s, _, hv, fieldsOf_cons.mpr ⟨s', fs₀, hv', h₀, rfl⟩, rfl⟩,
      List.Perm.swap _ _ _⟩
  | trans _ _ ih₁ ih₂ =>
    intro fs hf
    obtain ⟨fs₁, h₁, p₁⟩ := ih₁ hf
    obtain ⟨fs₂, h₂, p₂⟩ := ih₂ h₁
    exact ⟨fs₂, h₂, p₁.trans p₂⟩

/-- Key order is not shape: an object's members in any order have one
shape. -/
theorem shapeOf_key_order {kvs kvs' : List (String × Value)} (h : kvs.Perm kvs') :
    shapeOf (.obj kvs) = shapeOf (.obj kvs') := by
  have hk : (kvs.map (·.1)).Nodup ↔ (kvs'.map (·.1)).Nodup := (h.map _).nodup_iff
  simp only [shapeOf]
  by_cases nd : (kvs.map (·.1)).Nodup
  · rw [ite_eq_left nd, ite_eq_left (hk.mp nd)]
    cases h₁ : fieldsOf kvs with
    | none =>
      cases h₂ : fieldsOf kvs' with
      | none => rfl
      | some fs' => obtain ⟨fs, h₃, -⟩ := fieldsOf_perm h.symm h₂; rw [h₁] at h₃; cases h₃
    | some fs =>
      obtain ⟨fs', h₂, p⟩ := fieldsOf_perm h h₁
      rw [h₂]
      simp only [Option.map_some, Function.comp, Option.some.injEq, Shape.obj.injEq]
      exact sort_unique (f := (·.1)) (fun _ _ => rfl) p (by rw [fieldsOf_keys h₁]; exact nd)
  · rw [ite_eq_right nd, ite_eq_right (fun h' => nd (hk.mpr h'))]

/-- A repeated key has no shape: it is refused, never resolved. -/
theorem shapeOf_dup {kvs : List (String × Value)} (h : ¬ (kvs.map (·.1)).Nodup) :
    shapeOf (.obj kvs) = none := by
  simp [shapeOf, h]

theorem shapeOf_obj {kvs : List (String × Value)} {s} (h : shapeOf (.obj kvs) = some s) :
    (kvs.map (·.1)).Nodup ∧ ∃ fs, fieldsOf kvs = some fs ∧ s = .obj (sortKeys fs) := by
  simp only [shapeOf] at h
  split at h
  · rename_i nd
    cases h₁ : fieldsOf kvs with
    | none => rw [h₁] at h; cases h
    | some fs => rw [h₁] at h; simp only [Option.map_some, Function.comp, Option.some.injEq] at h; exact ⟨nd, fs, rfl, h.symm⟩
  · cases h

/-- Key sets are strict: objects with different keys have different
shapes. -/
theorem shapeOf_key_set {kvs kvs' : List (String × Value)} (h₁ : shapeOf (.obj kvs) = some s₁)
    (h₂ : shapeOf (.obj kvs') = some s₂) (hk : k ∈ kvs.map (·.1)) (hk' : k ∉ kvs'.map (·.1)) :
    s₁ ≠ s₂ := by
  obtain ⟨-, fs, hf, rfl⟩ := shapeOf_obj h₁
  obtain ⟨-, fs', hf', rfl⟩ := shapeOf_obj h₂
  intro e
  simp only [Shape.obj.injEq] at e
  apply hk'
  rw [← fieldsOf_keys hf']
  have m : k ∈ (sortKeys fs).map (·.1) := (sortKeys_perm.map _).symm.subset (by rw [fieldsOf_keys hf]; exact hk)
  rw [e] at m
  exact (sortKeys_perm.map _).subset m

/-- A key whose value has another shape gives another shape, at any
depth. -/
theorem shapeOf_key_kind {kvs kvs' : List (String × Value)} (h₁ : shapeOf (.obj kvs) = some s₁)
    (h₂ : shapeOf (.obj kvs') = some s₂) (hv : (k, v) ∈ kvs) (hv' : (k, v') ∈ kvs')
    (hne : shapeOf v ≠ shapeOf v') : s₁ ≠ s₂ := by
  obtain ⟨nd, fs, hf, rfl⟩ := shapeOf_obj h₁
  obtain ⟨-, fs', hf', rfl⟩ := shapeOf_obj h₂
  obtain ⟨t, ht, hm⟩ := fieldsOf_mem hf hv
  obtain ⟨t', ht', hm'⟩ := fieldsOf_mem hf' hv'
  intro e
  simp only [Shape.obj.injEq] at e
  have m : (k, t) ∈ sortKeys fs := sortKeys_perm.symm.subset hm
  have m' : (k, t') ∈ sortKeys fs := by rw [e]; exact sortKeys_perm.symm.subset hm'
  have ndk : ((sortKeys fs).map (·.1)).Nodup :=
    (sortKeys_perm.map _).nodup_iff.mpr (by rw [fieldsOf_keys hf]; exact nd)
  have := inj_of_nodup_map ndk m m' rfl
  simp only [Prod.mk.injEq] at this
  exact hne (by rw [ht, ht', this.2])

/-! ## Arrival -/

mutual
/-- A value JSON can express: no `bytes` or exit status anywhere in it. -/
def Value.isJson : Value → Bool
  | .bytes _ => false
  | .exit _ => false
  | .arr xs => allJson xs
  | .obj kvs => allJsonMembers kvs
  | _ => true

def allJson : List Value → Bool
  | [] => true
  | x :: xs => x.isJson && allJson xs

def allJsonMembers : List (String × Value) → Bool
  | [] => true
  | (_, v) :: kvs => v.isJson && allJsonMembers kvs
end

theorem allJsonMembers_mem : ∀ {kvs : List (String × Value)}, allJsonMembers kvs = true →
    ∀ {k v}, (k, v) ∈ kvs → v.isJson = true
  | [], _, _, _, h => by simp at h
  | (k', v') :: kvs, h, k, v, hm => by
    simp only [allJsonMembers, Bool.and_eq_true] at h
    rcases List.mem_cons.mp hm with e | hm
    · simp only [Prod.mk.injEq] at e; obtain ⟨rfl, rfl⟩ := e; exact h.1
    · exact allJsonMembers_mem h.2 hm

theorem fieldsOf_mem_rev : ∀ {kvs : List (String × Value)} {fs}, fieldsOf kvs = some fs →
    ∀ {k s}, (k, s) ∈ fs → ∃ v, (k, v) ∈ kvs ∧ shapeOf v = some s
  | [], fs, h, _, _, hm => by simp [fieldsOf] at h; subst h; simp at hm
  | (k', v') :: kvs, fs, h, k, s, hm => by
    obtain ⟨s', fs₀, hv, hk, rfl⟩ := fieldsOf_cons.mp h
    rcases List.mem_cons.mp hm with e | hm
    · simp only [Prod.mk.injEq] at e; obtain ⟨rfl, rfl⟩ := e
      exact ⟨v', List.mem_cons_self .., hv⟩
    · obtain ⟨v, hv, e⟩ := fieldsOf_mem_rev hk hm; exact ⟨v, List.mem_cons_of_mem _ hv, e⟩

/-- Only bytes have the shape `bytes`. -/
theorem shapeOf_bytes_inv : ∀ {v : Value}, shapeOf v = some .bytes → ∃ b, v = .bytes b
  | .null, h | .bool _, h | .num _, h | .str _, h => by simp [shapeOf] at h
  | .bytes b, _ => ⟨b, rfl⟩
  | .exit _, h => by simp [shapeOf] at h
  | .arr xs, h => by
    simp only [shapeOf] at h
    cases e : shapesOf xs <;> rw [e] at h <;> simp [canonArr] at h
  | .obj kvs, h => by
    simp only [shapeOf] at h
    split at h
    · cases e : fieldsOf kvs <;> rw [e] at h <;> simp at h
    · cases h

/-! ### Where decoding fails -/

def contByte (b : UInt8) : Bool := 0x80 ≤ b && b ≤ 0xBF

/-- The length of the UTF-8 sequence at the head, or none if there is none:
no overlong forms, no surrogates, nothing past U+10FFFF. -/
def seqLen : List UInt8 → Option Nat
  | [] => none
  | b :: rest =>
    if b < 0x80 then some 1
    else if 0xC2 ≤ b && b ≤ 0xDF then
      match rest with
      | c :: _ => if contByte c then some 2 else none
      | _ => none
    else if 0xE0 ≤ b && b ≤ 0xEF then
      let lo : UInt8 := if b = 0xE0 then 0xA0 else 0x80
      let hi : UInt8 := if b = 0xED then 0x9F else 0xBF
      match rest with
      | c₁ :: c₂ :: _ => if lo ≤ c₁ && c₁ ≤ hi && contByte c₂ then some 3 else none
      | _ => none
    else if 0xF0 ≤ b && b ≤ 0xF4 then
      let lo : UInt8 := if b = 0xF0 then 0x90 else 0x80
      let hi : UInt8 := if b = 0xF4 then 0x8F else 0xBF
      match rest with
      | c₁ :: c₂ :: c₃ :: _ =>
        if lo ≤ c₁ && c₁ ≤ hi && contByte c₂ && contByte c₃ then some 4 else none
      | _ => none
    else none

def firstInvalidFrom : Nat → List UInt8 → Nat → Option Nat
  | 0, _, _ => none
  | fuel + 1, l, pos => match l with
    | [] => none
    | _ => match seqLen l with
      | none => some pos
      | some n => firstInvalidFrom fuel (l.drop n) (pos + n)

/-- The offset of the first byte that does not begin a UTF-8 sequence. -/
def firstInvalid (raw : ByteArray) : Option Nat := firstInvalidFrom raw.size raw.toList 0

/-- Why bytes did not decode. -/
inductive Failure where
  | notUtf8
  /-- Not one JSON value: its offset is the decoder's, the bytes read when it
  found the error. -/
  | syntax
  /-- A key repeated in one object: its offset is just past the repeat. -/
  | duplicate (key : String)
  deriving Repr, DecidableEq

/-- What failed, where, and an example of what was expected. -/
def Failure.message (off : Nat) : Failure → String
  | .notUtf8 => s!"not UTF-8 at byte {off}: text is UTF-8, e.g. é is the two bytes c3 a9"
  | .syntax => s!"not JSON at byte {off}: the output is one JSON value, e.g. " ++
      "{\"name\": \"a\", \"size\": 3}"
  | .duplicate k => s!"not JSON at byte {off}: the key \"" ++ escape k ++
      "\" appears twice; each key appears once, e.g. {\"a\": 1, \"b\": 2}"

/-- What did not decode: the bytes themselves, untouched; where the first
error is; and what was expected. -/
def errorValue (raw : ByteArray) (off : Nat) (f : Failure) : Value :=
  .obj [("error", .bytes raw), ("at", .num (toString off)), ("message", .str (f.message off))]

def errorShape : Shape := .obj [("at", .num), ("error", .bytes), ("message", .str)]

theorem errorValue_shape : shapeOf (errorValue raw off f) = some errorShape := by
  simp [errorValue, errorShape, shapeOf, fieldsOf, sortKeys, List.Nodup, keyLe, List.mergeSort]

/-- No value JSON can express has the error's shape: a tool printing
`{"error": …}` is never mistaken for output that did not decode. -/
theorem json_no_bytes_field {v : Value} (hj : v.isJson = true) {fs : List (String × Shape)}
    (h : shapeOf v = some (.obj fs)) {k : String} (m : (k, Shape.bytes) ∈ fs) : False := by
  cases v with
  | null | bool _ | num _ | str _ => simp [shapeOf] at h
  | bytes b => simp [Value.isJson] at hj
  | exit _ => simp [Value.isJson] at hj
  | arr xs =>
    simp only [shapeOf] at h
    cases e : shapesOf xs <;> rw [e] at h <;> simp [canonArr] at h
  | obj kvs =>
    obtain ⟨-, fs', hf, e⟩ := shapeOf_obj h
    simp only [Shape.obj.injEq] at e
    have m' : (k, Shape.bytes) ∈ fs' := sortKeys_perm.subset (by rw [← e]; exact m)
    obtain ⟨v, hv, hs⟩ := fieldsOf_mem_rev hf m'
    obtain ⟨b, rfl⟩ := shapeOf_bytes_inv hs
    have := allJsonMembers_mem (by simpa [Value.isJson] using hj) hv
    simp [Value.isJson] at this

theorem json_not_error {v : Value} (hj : v.isJson = true) : shapeOf v ≠ some errorShape :=
  fun h => json_no_bytes_field hj h (k := "error") (by simp)

/-! ### A process that failed -/

/-- What a process that exited non-zero gives: its status, and what it
wrote to stdout and stderr, as bytes, untouched; a failing process's
output need not follow its contract. -/
def failedValue (exit : Nat) (stdout stderr : ByteArray) : Value :=
  .obj [("exit", .num (toString exit)), ("stderr", .bytes stderr), ("stdout", .bytes stdout)]

def failedShape : Shape := .obj [("exit", .num), ("stderr", .bytes), ("stdout", .bytes)]

theorem failedValue_shape : shapeOf (failedValue exit out err) = some failedShape := by
  simp [failedValue, failedShape, shapeOf, fieldsOf, sortKeys, List.Nodup, keyLe, List.mergeSort]

/-- No value JSON can express has a failure's shape. -/
theorem json_not_failed {v : Value} (hj : v.isJson = true) : shapeOf v ≠ some failedShape :=
  fun h => json_no_bytes_field hj h (k := "stdout") (by simp)

/-- A failure is not a decode error: they are different bases. -/
theorem failed_not_error : failedShape ≠ errorShape := by
  simp [failedShape, errorShape]

/-! ### A frame's return -/

/-- What a frame's `return` gives its caller, built by the core, never
written: its exit status, 0 when the goal is done and 1 when it is not, and
what it returns as bytes, opaque until something runs over them; stderr
only when the return names one. A tool's output is the OS's and is not
this: only a frame, which the core makes, returns it. -/
def returnValue (exit : Nat) (stdout : ByteArray) : Option ByteArray → Value
  | none => .obj [("exit", .exit exit), ("stdout", .bytes stdout)]
  | some err => .obj [("exit", .exit exit), ("stderr", .bytes err), ("stdout", .bytes stdout)]

def returnShape (exit : Nat) : Bool → Shape
  | false => .obj [("exit", .exit exit), ("stdout", .bytes)]
  | true => .obj [("exit", .exit exit), ("stderr", .bytes), ("stdout", .bytes)]

theorem returnValue_shape : shapeOf (returnValue exit out err) = some (returnShape exit err.isSome) := by
  cases err <;> simp [returnValue, returnShape, shapeOf, fieldsOf, sortKeys, List.Nodup, keyLe, List.mergeSort]

/-- No value JSON can express has a return's shape: a value a tool or a
child printed is never taken for a return. -/
theorem json_not_return {v : Value} (hj : v.isJson = true) : shapeOf v ≠ some (returnShape exit e) := by
  intro h
  cases e
  · exact json_no_bytes_field hj h (k := "stdout") (by simp)
  · exact json_no_bytes_field hj h (k := "stdout") (by simp)

/-- Two exit statuses are two shapes: the runtime tells them apart before
any guard runs. -/
theorem returnShape_exit (h : x ≠ y) : returnShape x e ≠ returnShape y e' := by
  cases e <;> cases e' <;> simp [returnShape, h]

/-- A return is not a process's failure: its exit is a status, a failure's
a number. -/
theorem return_not_failed : returnShape exit e ≠ failedShape := by
  cases e <;> simp [returnShape, failedShape]

/-- A contract's declared output. Without a declaration it is `text`. -/
inductive Output where
  | json
  | text
  deriving Repr, DecidableEq

/-- What arrives from a contract's output bytes. Bytes that are not UTF-8
are the error, under either declaration. Under `text`, the bytes as a
string; under `json`, the value they decode to (`decoded` is the shell's
JSON decoding, which Lean does not do: the value, or where and why it
failed). -/
def arrive (o : Output) (raw : ByteArray) (decoded : Except (Nat × Failure) Value) : Value :=
  match firstInvalid raw with
  | some off => errorValue raw off .notUtf8
  | none => match o with
    | .text => match String.fromUTF8? raw with
      | some s => .str s
      | none => errorValue raw 0 .notUtf8
    | .json => match decoded with
      | .ok v => if v.isJson && (shapeOf v).isSome then v else errorValue raw 0 .syntax
      | .error (off, f) => errorValue raw off f

/-- Under `text`, UTF-8 output is a string, whatever it looks like: `5`
from `wc -l` is the string, even though it decodes as JSON. -/
theorem arrive_text (hv : firstInvalid raw = none) (h : String.fromUTF8? raw = some s)
    (d : Except (Nat × Failure) Value) : arrive .text raw d = .str s := by
  simp [arrive, hv, h]

/-- Every arrival has a shape: output that does not decode is the error. -/
theorem arrive_shaped (o : Output) (raw : ByteArray) (d : Except (Nat × Failure) Value) :
    (shapeOf (arrive o raw d)).isSome := by
  unfold arrive
  cases hf : firstInvalid raw with
  | some off => simp [errorValue_shape]
  | none =>
    cases o with
    | text => cases hu : String.fromUTF8? raw <;> simp [shapeOf, errorValue_shape]
    | json =>
      cases d with
      | error e => obtain ⟨off, f⟩ := e; simp [errorValue_shape]
      | ok v =>
        by_cases hc : v.isJson = true ∧ (shapeOf v).isSome = true
        · simp [hc]
        · simp [hc, errorValue_shape]

/-- An arrival has the error's shape exactly when it did not decode: never
for what decoded. -/
theorem arrive_error (o : Output) (raw : ByteArray) (d : Except (Nat × Failure) Value) :
    shapeOf (arrive o raw d) = some errorShape ↔ ∃ off f, arrive o raw d = errorValue raw off f := by
  constructor
  · intro h
    unfold arrive at h ⊢
    cases hf : firstInvalid raw with
    | some off => exact ⟨off, .notUtf8, by simp⟩
    | none =>
      rw [hf] at h
      cases o with
      | text =>
        cases hu : String.fromUTF8? raw with
        | none => exact ⟨0, .notUtf8, by simp⟩
        | some s => simp [hu, shapeOf, errorShape] at h
      | json =>
        cases d with
        | error e => obtain ⟨off, f⟩ := e; exact ⟨off, f, by simp⟩
        | ok v =>
          by_cases hc : v.isJson = true ∧ (shapeOf v).isSome = true
          · simp [hc] at h
            exact absurd h (json_not_error hc.1)
          · exact ⟨0, .syntax, by simp [hc]⟩
  · rintro ⟨off, f, h⟩; rw [h]; exact errorValue_shape

/-- What arrives from a process: exiting 0, its stdout read as its
contract declares (`arrive`); otherwise the failure. -/
def arriveExit (o : Output) (stdout : ByteArray) (decoded : Except (Nat × Failure) Value)
    (exit : Nat) (stderr : ByteArray) : Value :=
  if exit = 0 then arrive o stdout decoded else failedValue exit stdout stderr

/-- Exiting 0 changes nothing: the success path is `arrive`'s. -/
theorem arriveExit_zero : arriveExit o out d 0 err = arrive o out d := by
  simp [arriveExit]

theorem arriveExit_shaped (o : Output) (out : ByteArray) (d : Except (Nat × Failure) Value)
    (exit : Nat) (err : ByteArray) : (shapeOf (arriveExit o out d exit err)).isSome := by
  unfold arriveExit
  split
  · exact arrive_shaped o out d
  · simp [failedValue_shape]

/-- An arrival has a failure's shape exactly when the process exited
non-zero: never for what a process that succeeded printed. -/
theorem arriveExit_failed (o : Output) (out : ByteArray) (d : Except (Nat × Failure) Value)
    (exit : Nat) (err : ByteArray) :
    shapeOf (arriveExit o out d exit err) = some failedShape ↔ exit ≠ 0 := by
  unfold arriveExit
  split
  · rename_i h0
    simp only [h0, ne_eq, not_true_eq_false, iff_false]
    intro h
    by_cases he : shapeOf (arrive o out d) = some errorShape
    · rw [he] at h; exact failed_not_error (Option.some.inj h).symm
    · unfold arrive at h he
      cases hf : firstInvalid out with
      | some off => simp [hf, errorValue_shape] at he
      | none =>
        simp only [hf] at h he
        cases o with
        | text =>
          cases hu : String.fromUTF8? out with
          | none => simp [hu, errorValue_shape] at he
          | some s => simp [hu, shapeOf, failedShape] at h
        | json =>
          cases d with
          | error e => obtain ⟨off, f⟩ := e; simp [errorValue_shape] at he
          | ok v =>
            by_cases hc : v.isJson = true ∧ (shapeOf v).isSome = true
            · simp [hc] at h; exact json_not_failed hc.1 h
            · simp [hc, errorValue_shape] at he
  · rename_i h0; simp [h0, failedValue_shape]

end Shape
